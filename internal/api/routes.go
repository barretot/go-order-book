package api

import (
	"github.com/barretot/go-order-book/internal/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handlers.Handlers) {

	r.Use(
		gin.Logger(),
		gin.Recovery(),
	)

	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			users := v1.Group("/users")
			{
				users.POST("", h.User.HandleCreateUser)
				users.DELETE("/:userId/orders/:orderId", h.Order.HandleCancelOrder)
			}

			wallets := v1.Group("/wallets")
			{
				wallets.POST("/:id", h.Wallet.HandleAddWalletAsset)
				wallets.GET("/:id", h.Wallet.HandleGetWallet)
			}

			orders := v1.Group("/orders")
			{
				orders.POST("/:userId", h.Order.HandlePlaceOrder)
			}

			orderBook := v1.Group("/orderbook")
			{
				orderBook.GET("", h.Order.HandleGetOrderBook)
			}
		}
	}

}
