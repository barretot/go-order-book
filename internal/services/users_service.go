package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/barretot/go-order-book/internal/domain/models"
	"github.com/barretot/go-order-book/internal/store/pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDuplicatedEmail    = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidCode        = errors.New("invalid or expired code")
)

type UserService struct {
	pool    *pgxpool.Pool
	queries *pg.Queries
}

func NewUserService(pool *pgxpool.Pool) *UserService {
	return &UserService{
		pool:    pool,
		queries: pg.New(pool),
	}
}

func (us *UserService) CreateUser(ctx context.Context, input models.User) (uuid.UUID, error) {

	args := pg.CreateUserParams{
		Name:  input.Name,
		Email: input.Email,
	}

	user_id, err := us.queries.CreateUser(ctx, args)
	if err != nil {
		var pgErr *pgconn.PgError
		// erro 23505 = unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			slog.Warn("duplicate constraint violation",
				"constraint", pgErr.ConstraintName,
				"email", input.Email,
			)
			return uuid.UUID{}, ErrDuplicatedEmail
		}
		slog.Error("database error creating user",
			"error", err,
			"email", input.Email,
		)
		return uuid.UUID{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user_id, nil
}
