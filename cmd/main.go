package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/barretot/go-order-book/internal/api"
	"github.com/barretot/go-order-book/internal/handlers"
	"github.com/barretot/go-order-book/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	slog.Info("starting application", "service", "go-order-book")

	if err := godotenv.Load(); err != nil {
		slog.Error("failed to load .env file", "error", err)
		log.Fatal("Failed to load .env:", err)
	}

	slog.Info("environment variables loaded successfully")

	ctx := context.Background()

	slog.Info("connecting to database",
		"host", os.Getenv("DATABASE_HOST"),
		"port", os.Getenv("DATABASE_PORT"),
		"database", os.Getenv("DATABASE_NAME"),
	)

	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_NAME"),
	))

	if err != nil {
		slog.Error("failed to connect to database",
			"error", err,
			"host", os.Getenv("DATABASE_HOST"),
			"port", os.Getenv("DATABASE_PORT"),
		)
		log.Fatal("Failed to connect to database:", err)
	}

	ordersService := services.NewOrdersService(pool)
	userService := services.NewUserService(pool)
	walletService := services.NewWalletService(pool)

	h := &handlers.Handlers{
		Order: &handlers.OrdersHandler{
			OrdersService: ordersService,
		},
		User: &handlers.UserHandler{
			UserService: userService,
		},
		Wallet: &handlers.WalletHandler{
			WalletService: walletService,
		},
	}

	slog.Info("setting up HTTP router")
	r := gin.Default()

	api.RegisterRoutes(r, h)
	slog.Info("routes registered successfully")

	port := ":8080"
	slog.Info("starting HTTP server", "port", port)
	fmt.Println("Server running on", port)
	if err := r.Run(port); err != nil {
		slog.Error("failed to start server", "error", err, "port", port)
		log.Fatal("Failed to start server:", err)
	}
}
