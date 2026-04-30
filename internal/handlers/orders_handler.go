package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/barretot/go-order-book/internal/apperrors"
	"github.com/barretot/go-order-book/internal/domain/dto"
	"github.com/barretot/go-order-book/internal/domain/models"
	"github.com/barretot/go-order-book/internal/services"
	"github.com/barretot/go-order-book/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type OrdersHandler struct {
	OrdersService *services.OrdersService
}

func (h *OrdersHandler) HandlePlaceOrder(c *gin.Context) {
	var request dto.PlaceOrderRequest
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			errors := make(map[string]string)
			for _, fieldErr := range verrs {
				errors[fieldErr.Field()] = utils.GetErrorMessage(fieldErr)
			}
			slog.Error("validation request error", "error", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"status":         "denied",
				"reason":         errors,
				"accepted_sides": models.AcceptedSides,
			})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid payload"})
		return
	}

	input := models.Order{
		UserID:     userID,
		Instrument: models.Instrument(request.Instrument),
		Side:       models.Side(request.Side),
		Price:      request.Price,
		Quantity:   request.Quantity,
	}

	result, err := h.OrdersService.CreateOrder(c.Request.Context(), input)
	var validationError = &apperrors.ValidationError{}

	if errors.As(err, &validationError) {
		slog.Error("order validation error", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "denied",
			"reason": err.Error(),
		})
		return
	}

	if err != nil {
		slog.Error("failed to create order", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	slog.Info("order accepted")
	c.JSON(http.StatusOK, gin.H{
		"order_id":           result.ID,
		"user_id":            result.UserID,
		"instrument":         result.Instrument,
		"quantity":           result.Quantity,
		"remaining_quantity": result.RemainingQuantity,
		"price":              result.Price,
		"side":               result.Side,
		"status":             result.Status,
	})
}

func (h *OrdersHandler) HandleCancelOrder(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	result, err := h.OrdersService.CancelOrder(c.Request.Context(), userID, orderID)
	var validationError = &apperrors.ValidationError{}

	if errors.As(err, &validationError) {
		slog.Error("order cancellation validation error", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "denied",
			"reason": err.Error(),
		})
		return
	}

	if err != nil {
		slog.Error("failed to cancel order", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel order"})
		return
	}

	slog.Info("order cancelled")
	c.JSON(http.StatusOK, gin.H{
		"order_id":           result.ID,
		"user_id":            result.UserID,
		"instrument":         result.Instrument,
		"quantity":           result.Quantity,
		"remaining_quantity": result.RemainingQuantity,
		"price":              result.Price,
		"side":               result.Side,
		"status":             result.Status,
	})
}

func (h *OrdersHandler) HandleGetOrderBook(c *gin.Context) {
	instrument := c.Query("instrument")
	if instrument == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instrument is required"})
		return
	}

	orderBook, err := h.OrdersService.GetOrderBook(c.Request.Context(), models.Instrument(instrument))
	if err != nil {
		slog.Error("failed to get order book", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order book"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instrument": orderBook.Instrument,
		"bids":       orderBook.Bids,
		"asks":       orderBook.Asks,
	})
}
