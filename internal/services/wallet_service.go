package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/barretot/go-order-book/internal/domain/models"
	"github.com/barretot/go-order-book/internal/store/pg"
	"github.com/barretot/go-order-book/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletService struct {
	pool    *pgxpool.Pool
	queries *pg.Queries
}

func NewWalletService(pool *pgxpool.Pool) *WalletService {
	return &WalletService{
		pool:    pool,
		queries: pg.New(pool),
	}
}

func (us *WalletService) AddWalletAsset(ctx context.Context, input models.WalletAssets) (uuid.UUID, error) {

	args := pg.CreateWalletAssetParams{
		UserID: input.UserId,
		Instrument: pgtype.Text{
			String: string(input.Instrument),
			Valid:  input.Instrument != "",
		},
		Quantity: utils.NumericFromFloat(input.Quantity),
	}

	wallet_id, err := us.queries.CreateWalletAsset(ctx, args)

	if err != nil {
		slog.Error("database error creating wallet", "error", err)
		return uuid.UUID{}, fmt.Errorf("failed to create wallet asset: %w", err)
	}

	return wallet_id, nil
}

func (us *WalletService) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (models.WalletDBModel, error) {
	user, err := us.queries.GetUserById(ctx, userID)
	if err != nil {
		slog.Error("database error getting user", "error", err)
		return models.WalletDBModel{}, fmt.Errorf("failed to get user: %w", err)
	}

	walletAssets, err := us.queries.GetWalletByUserId(ctx, userID)
	if err != nil {
		slog.Error("database error getting wallet assets", "error", err)
		return models.WalletDBModel{}, fmt.Errorf("failed to get wallet assets: %w", err)
	}

	result := make([]models.WalletAssetDBModel, 0, len(walletAssets))
	for _, asset := range walletAssets {
		quantity, err := asset.Quantity.Float64Value()
		if err != nil {
			return models.WalletDBModel{}, fmt.Errorf("failed to parse wallet asset quantity: %w", err)
		}

		result = append(result, models.WalletAssetDBModel{
			ID:         asset.ID,
			UserID:     asset.UserID,
			Instrument: models.Instrument(asset.Instrument.String),
			Quantity:   quantity.Float64,
		})
	}

	return models.WalletDBModel{
		UserID: user.ID,
		Email:  user.Email,
		Assets: result,
	}, nil
}
