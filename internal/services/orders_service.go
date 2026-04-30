package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/barretot/go-order-book/internal/apperrors"
	"github.com/barretot/go-order-book/internal/domain/models"
	"github.com/barretot/go-order-book/internal/store/pg"
	"github.com/barretot/go-order-book/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrdersService struct {
	pool    *pgxpool.Pool
	Queries *pg.Queries
}

func NewOrdersService(pool *pgxpool.Pool) *OrdersService {
	return &OrdersService{
		pool:    pool,
		Queries: pg.New(pool),
	}
}

func (us *OrdersService) CreateOrder(ctx context.Context, input models.Order) (models.OrderDBModel, error) {
	baseInstrument, quoteInstrument := utils.ParseInstrumentPair(input.Instrument)

	if input.Side == models.Sell {
		if err := us.ensureWalletHas(ctx, input.UserID, baseInstrument, input.Quantity); err != nil {
			return models.OrderDBModel{}, err
		}
	}

	if input.Side == models.Buy {
		if err := us.ensureWalletHas(ctx, input.UserID, quoteInstrument, input.Quantity*input.Price); err != nil {
			return models.OrderDBModel{}, err
		}
	}

	args := pg.CreateOrderParams{
		UserID:     input.UserID,
		Instrument: string(input.Instrument),
		Quantity:   utils.NumericFromFloat(input.Quantity),
		Price:      utils.NumericFromFloat(input.Price),
		Side:       string(input.Side),
	}

	tx, err := us.pool.Begin(ctx)
	if err != nil {
		return models.OrderDBModel{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := us.Queries.WithTx(tx)
	result, err := qtx.CreateOrder(ctx, args)
	if err != nil {
		slog.Error("database error creating order", "error", err)
		return models.OrderDBModel{}, fmt.Errorf("failed to create order: %w", err)
	}

	order := orderDBModelFromCreateOrder(result)
	for order.RemainingQuantity > 0 {
		match, found, err := us.findMatch(ctx, qtx, input)
		if err != nil {
			return models.OrderDBModel{}, err
		}
		if !found {
			break
		}

		tradeQuantity := utils.MinFloat(order.RemainingQuantity, match.remainingQuantity)
		if err := us.settleTrade(ctx, qtx, input, order.ID, match, baseInstrument, quoteInstrument, tradeQuantity); err != nil {
			return models.OrderDBModel{}, err
		}

		updatedOrder, err := qtx.DecrementOrderRemaining(ctx, pg.DecrementOrderRemainingParams{
			ID:                order.ID,
			RemainingQuantity: utils.NumericFromFloat(tradeQuantity),
		})
		if err != nil {
			return models.OrderDBModel{}, fmt.Errorf("failed to update order remaining quantity: %w", err)
		}

		if _, err := qtx.DecrementOrderRemaining(ctx, pg.DecrementOrderRemainingParams{
			ID:                match.id,
			RemainingQuantity: utils.NumericFromFloat(tradeQuantity),
		}); err != nil {
			return models.OrderDBModel{}, fmt.Errorf("failed to update matched order remaining quantity: %w", err)
		}

		order = orderDBModelFromDecrementOrder(updatedOrder)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.OrderDBModel{}, fmt.Errorf("failed to commit order transaction: %w", err)
	}

	return order, nil
}

func (us *OrdersService) CancelOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (models.OrderDBModel, error) {
	order, err := us.Queries.CancelOrder(ctx, pg.CancelOrderParams{
		ID:     orderID,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return models.OrderDBModel{}, &apperrors.ValidationError{
			Status: "denied",
			Reason: "order not found, does not belong to user, or cannot be cancelled",
		}
	}
	if err != nil {
		return models.OrderDBModel{}, fmt.Errorf("failed to cancel order: %w", err)
	}

	return orderDBModelFromCancelOrder(order), nil
}

func (us *OrdersService) GetOrderBook(ctx context.Context, instrument models.Instrument) (models.OrderBook, error) {
	orders, err := us.Queries.GetOrderBookByInstrument(ctx, string(instrument))
	if err != nil {
		return models.OrderBook{}, fmt.Errorf("failed to get order book: %w", err)
	}

	orderBook := models.OrderBook{
		Instrument: instrument,
		Bids:       make([]models.OrderDBModel, 0),
		Asks:       make([]models.OrderDBModel, 0),
	}

	for _, order := range orders {
		orderModel := orderDBModelFromOrderBook(order)
		if orderModel.Side == models.Buy {
			orderBook.Bids = append(orderBook.Bids, orderModel)
			continue
		}

		orderBook.Asks = append(orderBook.Asks, orderModel)
	}

	return orderBook, nil
}

type matchingOrder struct {
	id                uuid.UUID
	userID            uuid.UUID
	remainingQuantity float64
	price             float64
	side              models.Side
}

func (us *OrdersService) ensureWalletHas(ctx context.Context, userID uuid.UUID, instrument models.Instrument, requiredQuantity float64) error {
	walletAssets, err := us.Queries.GetWalletByUserId(ctx, userID)
	if err != nil {
		slog.Error("database error getting wallet assets", "error", err)
		return fmt.Errorf("failed to get wallet assets: %w", err)
	}

	for _, asset := range walletAssets {
		if !asset.Instrument.Valid || asset.Instrument.String != string(instrument) {
			continue
		}

		quantity, err := asset.Quantity.Float64Value()
		if err != nil {
			return fmt.Errorf("failed to parse wallet asset quantity: %w", err)
		}

		if quantity.Float64 >= requiredQuantity {
			return nil
		}
	}

	return &apperrors.ValidationError{
		Status: "denied",
		Reason: fmt.Sprintf("insufficient quantity for instrument %s", instrument),
	}
}

func (us *OrdersService) findMatch(ctx context.Context, qtx *pg.Queries, input models.Order) (matchingOrder, bool, error) {
	if input.Side == models.Buy {
		match, err := qtx.FindSellMatchForBuy(ctx, pg.FindSellMatchForBuyParams{
			Instrument: string(input.Instrument),
			Price:      utils.NumericFromFloat(input.Price),
			UserID:     input.UserID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return matchingOrder{}, false, nil
		}
		if err != nil {
			return matchingOrder{}, false, fmt.Errorf("failed to find sell match: %w", err)
		}

		remaining, err := match.RemainingQuantity.Float64Value()
		if err != nil {
			return matchingOrder{}, false, fmt.Errorf("failed to parse matched order remaining quantity: %w", err)
		}
		price, err := match.Price.Float64Value()
		if err != nil {
			return matchingOrder{}, false, fmt.Errorf("failed to parse matched order price: %w", err)
		}

		return matchingOrder{
			id:                match.ID,
			userID:            match.UserID,
			remainingQuantity: remaining.Float64,
			price:             price.Float64,
			side:              models.Side(match.Side),
		}, true, nil
	}

	match, err := qtx.FindBuyMatchForSell(ctx, pg.FindBuyMatchForSellParams{
		Instrument: string(input.Instrument),
		Price:      utils.NumericFromFloat(input.Price),
		UserID:     input.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return matchingOrder{}, false, nil
	}
	if err != nil {
		return matchingOrder{}, false, fmt.Errorf("failed to find buy match: %w", err)
	}

	remaining, err := match.RemainingQuantity.Float64Value()
	if err != nil {
		return matchingOrder{}, false, fmt.Errorf("failed to parse matched order remaining quantity: %w", err)
	}
	price, err := match.Price.Float64Value()
	if err != nil {
		return matchingOrder{}, false, fmt.Errorf("failed to parse matched order price: %w", err)
	}

	return matchingOrder{
		id:                match.ID,
		userID:            match.UserID,
		remainingQuantity: remaining.Float64,
		price:             price.Float64,
		side:              models.Side(match.Side),
	}, true, nil
}

func (us *OrdersService) settleTrade(
	ctx context.Context,
	qtx *pg.Queries,
	input models.Order,
	orderID uuid.UUID,
	match matchingOrder,
	baseInstrument models.Instrument,
	quoteInstrument models.Instrument,
	tradeQuantity float64,
) error {
	tradePrice := match.price
	quoteQuantity := tradeQuantity * tradePrice

	buyOrderID := orderID
	sellOrderID := match.id
	buyerID := input.UserID
	sellerID := match.userID

	if input.Side == models.Sell {
		buyOrderID = match.id
		sellOrderID = orderID
		buyerID = match.userID
		sellerID = input.UserID
	}

	if err := debitWallet(ctx, qtx, buyerID, quoteInstrument, quoteQuantity); err != nil {
		return err
	}
	if err := creditWallet(ctx, qtx, buyerID, baseInstrument, tradeQuantity); err != nil {
		return err
	}
	if err := debitWallet(ctx, qtx, sellerID, baseInstrument, tradeQuantity); err != nil {
		return err
	}
	if err := creditWallet(ctx, qtx, sellerID, quoteInstrument, quoteQuantity); err != nil {
		return err
	}

	if _, err := qtx.CreateTrade(ctx, pg.CreateTradeParams{
		BuyOrderID:  buyOrderID,
		SellOrderID: sellOrderID,
		Instrument:  string(input.Instrument),
		Quantity:    utils.NumericFromFloat(tradeQuantity),
		Price:       utils.NumericFromFloat(tradePrice),
	}); err != nil {
		return fmt.Errorf("failed to create trade: %w", err)
	}

	return nil
}

func debitWallet(ctx context.Context, qtx *pg.Queries, userID uuid.UUID, instrument models.Instrument, quantity float64) error {
	rowsAffected, err := qtx.DebitWalletAsset(ctx, pg.DebitWalletAssetParams{
		UserID: userID,
		Instrument: pgtype.Text{
			String: string(instrument),
			Valid:  true,
		},
		Quantity: utils.NumericFromFloat(quantity),
	})
	if err != nil {
		return fmt.Errorf("failed to debit wallet asset: %w", err)
	}
	if rowsAffected == 0 {
		return &apperrors.ValidationError{
			Status: "denied",
			Reason: fmt.Sprintf("insufficient quantity for instrument %s", instrument),
		}
	}

	return nil
}

func creditWallet(ctx context.Context, qtx *pg.Queries, userID uuid.UUID, instrument models.Instrument, quantity float64) error {
	_, err := qtx.CreateWalletAsset(ctx, pg.CreateWalletAssetParams{
		UserID: userID,
		Instrument: pgtype.Text{
			String: string(instrument),
			Valid:  true,
		},
		Quantity: utils.NumericFromFloat(quantity),
	})
	if err != nil {
		return fmt.Errorf("failed to credit wallet asset: %w", err)
	}

	return nil
}

func orderDBModelFromCreateOrder(order pg.CreateOrderRow) models.OrderDBModel {
	quantity, _ := order.Quantity.Float64Value()
	remainingQuantity, _ := order.RemainingQuantity.Float64Value()
	price, _ := order.Price.Float64Value()

	return models.OrderDBModel{
		ID:                order.ID,
		UserID:            order.UserID,
		Instrument:        models.Instrument(order.Instrument),
		Quantity:          quantity.Float64,
		RemainingQuantity: remainingQuantity.Float64,
		Price:             price.Float64,
		Side:              models.Side(order.Side),
		Status:            order.Status,
	}
}

func orderDBModelFromDecrementOrder(order pg.DecrementOrderRemainingRow) models.OrderDBModel {
	quantity, _ := order.Quantity.Float64Value()
	remainingQuantity, _ := order.RemainingQuantity.Float64Value()
	price, _ := order.Price.Float64Value()

	return models.OrderDBModel{
		ID:                order.ID,
		UserID:            order.UserID,
		Instrument:        models.Instrument(order.Instrument),
		Quantity:          quantity.Float64,
		RemainingQuantity: remainingQuantity.Float64,
		Price:             price.Float64,
		Side:              models.Side(order.Side),
		Status:            order.Status,
	}
}

func orderDBModelFromCancelOrder(order pg.CancelOrderRow) models.OrderDBModel {
	quantity, _ := order.Quantity.Float64Value()
	remainingQuantity, _ := order.RemainingQuantity.Float64Value()
	price, _ := order.Price.Float64Value()

	return models.OrderDBModel{
		ID:                order.ID,
		UserID:            order.UserID,
		Instrument:        models.Instrument(order.Instrument),
		Quantity:          quantity.Float64,
		RemainingQuantity: remainingQuantity.Float64,
		Price:             price.Float64,
		Side:              models.Side(order.Side),
		Status:            order.Status,
	}
}

func orderDBModelFromOrderBook(order pg.Order) models.OrderDBModel {
	quantity, _ := order.Quantity.Float64Value()
	remainingQuantity, _ := order.RemainingQuantity.Float64Value()
	price, _ := order.Price.Float64Value()

	return models.OrderDBModel{
		ID:                order.ID,
		UserID:            order.UserID,
		Instrument:        models.Instrument(order.Instrument),
		Quantity:          quantity.Float64,
		RemainingQuantity: remainingQuantity.Float64,
		Price:             price.Float64,
		Side:              models.Side(order.Side),
		Status:            order.Status,
	}
}
