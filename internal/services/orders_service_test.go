package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/barretot/go-order-book/internal/apperrors"
	"github.com/barretot/go-order-book/internal/domain/models"
	"github.com/barretot/go-order-book/internal/store/pg"
	"github.com/barretot/go-order-book/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrdersServiceCancelOrder(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	db.ExpectQuery("UPDATE orders").
		WithArgs(orderID, userID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id",
			"user_id",
			"instrument",
			"quantity",
			"remaining_quantity",
			"price",
			"side",
			"status",
		}).AddRow(
			orderID,
			userID,
			"BTC/BRL",
			utils.NumericFromFloat(1),
			utils.NumericFromFloat(1),
			utils.NumericFromFloat(500000),
			"sell",
			"cancelled",
		))

	service := &OrdersService{Queries: pg.New(db)}

	order, err := service.CancelOrder(context.Background(), userID, orderID)

	require.NoError(t, err)
	assert.Equal(t, orderID, order.ID)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, models.Instrument("BTC/BRL"), order.Instrument)
	assert.Equal(t, models.Sell, order.Side)
	assert.Equal(t, "cancelled", order.Status)
	assert.Equal(t, 1.0, order.RemainingQuantity)
	assert.Equal(t, 500000.0, order.Price)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestOrdersServiceCancelOrderWhenOrderDoesNotBelongToUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	db.ExpectQuery("UPDATE orders").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(pgx.ErrNoRows)

	service := &OrdersService{Queries: pg.New(db)}

	_, err = service.CancelOrder(context.Background(), uuid.New(), uuid.New())

	var validationError *apperrors.ValidationError
	require.ErrorAs(t, err, &validationError)
	assert.Contains(t, validationError.Error(), "does not belong to user")
	require.NoError(t, db.ExpectationsWereMet())
}

func TestOrdersServiceGetOrderBookSplitsBidsAndAsks(t *testing.T) {
	buyOrderID := uuid.New()
	sellOrderID := uuid.New()
	buyerID := uuid.New()
	sellerID := uuid.New()
	now := time.Now()
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	db.ExpectQuery("FROM orders").
		WithArgs("BTC/BRL").
		WillReturnRows(pgxmock.NewRows([]string{
			"id",
			"user_id",
			"instrument",
			"quantity",
			"remaining_quantity",
			"price",
			"side",
			"status",
			"created_at",
			"updated_at",
		}).
			AddRow(
				buyOrderID,
				buyerID,
				"BTC/BRL",
				utils.NumericFromFloat(2),
				utils.NumericFromFloat(1),
				utils.NumericFromFloat(499000),
				"buy",
				"partially_filled",
				now,
				now,
			).
			AddRow(
				sellOrderID,
				sellerID,
				"BTC/BRL",
				utils.NumericFromFloat(1),
				utils.NumericFromFloat(1),
				utils.NumericFromFloat(501000),
				"sell",
				"open",
				now,
				now,
			))

	service := &OrdersService{Queries: pg.New(db)}

	orderBook, err := service.GetOrderBook(context.Background(), models.Instrument("BTC/BRL"))

	require.NoError(t, err)
	assert.Equal(t, models.Instrument("BTC/BRL"), orderBook.Instrument)
	require.Len(t, orderBook.Bids, 1)
	require.Len(t, orderBook.Asks, 1)
	assert.Equal(t, buyOrderID, orderBook.Bids[0].ID)
	assert.Equal(t, models.Buy, orderBook.Bids[0].Side)
	assert.Equal(t, sellOrderID, orderBook.Asks[0].ID)
	assert.Equal(t, models.Sell, orderBook.Asks[0].Side)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestOrdersServiceGetOrderBookReturnsQueryError(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	db.ExpectQuery("FROM orders").
		WithArgs("BTC/BRL").
		WillReturnError(assert.AnError)

	service := &OrdersService{Queries: pg.New(db)}

	_, err = service.GetOrderBook(context.Background(), models.Instrument("BTC/BRL"))

	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to get order book"))
	require.NoError(t, db.ExpectationsWereMet())
}
