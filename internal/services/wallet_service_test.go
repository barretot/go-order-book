package services

import (
	"context"
	"testing"
	"time"

	"github.com/barretot/go-order-book/internal/domain/models"
	"github.com/barretot/go-order-book/internal/store/pg"
	"github.com/barretot/go-order-book/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletServiceAddWalletAsset(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	db.ExpectQuery("INSERT INTO wallet_assets").
		WithArgs(userID, pgtype.Text{String: "BRL", Valid: true}, utils.NumericFromFloat(600000)).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(walletID))

	service := &WalletService{queries: pg.New(db)}

	result, err := service.AddWalletAsset(context.Background(), models.WalletAssets{
		UserId:     userID,
		Instrument: models.BRL,
		Quantity:   600000,
	})

	require.NoError(t, err)
	assert.Equal(t, walletID, result)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestWalletServiceGetWalletByUserID(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	now := time.Now()
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	db.ExpectQuery("FROM users").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id",
			"name",
			"email",
			"created_at",
			"updated_at",
		}).AddRow(
			userID,
			"Alice",
			"alice@example.com",
			pgtype.Timestamptz{Time: now, Valid: true},
			pgtype.Timestamptz{Time: now, Valid: true},
		))

	db.ExpectQuery("FROM wallet_assets").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id",
			"user_id",
			"instrument",
			"quantity",
			"created_at",
			"updated_at",
		}).AddRow(
			walletID,
			userID,
			pgtype.Text{String: "BRL", Valid: true},
			utils.NumericFromFloat(600000),
			now,
			now,
		))

	service := &WalletService{queries: pg.New(db)}

	wallet, err := service.GetWalletByUserID(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, userID, wallet.UserID)
	assert.Equal(t, "alice@example.com", wallet.Email)
	require.Len(t, wallet.Assets, 1)
	assert.Equal(t, walletID, wallet.Assets[0].ID)
	assert.Equal(t, models.BRL, wallet.Assets[0].Instrument)
	assert.Equal(t, 600000.0, wallet.Assets[0].Quantity)
	require.NoError(t, db.ExpectationsWereMet())
}
