package handlers

import (
	"log/slog"
	"net/http"

	"github.com/barretot/go-order-book/internal/domain/dto"
	"github.com/barretot/go-order-book/internal/domain/models"
	"github.com/barretot/go-order-book/internal/services"
	"github.com/barretot/go-order-book/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type WalletHandler struct {
	WalletService *services.WalletService
}

func (h *WalletHandler) HandleAddWalletAsset(c *gin.Context) {
	var request dto.AddWalletAssetRequest
	user_id := c.Param("id")
	userID, err := uuid.Parse(user_id)
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
				"reason":               errors,
				"accepted_instruments": models.AcceptedInstruments,
			})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid payload"})
		return
	}

	input := models.WalletAssets{
		UserId:     userID,
		Instrument: models.Instrument(request.Instrument),
		Quantity:   request.Quantity,
	}

	_, err = h.WalletService.AddWalletAsset(c.Request.Context(), input)
	if err != nil {
		slog.Error("failed to create wallet asset", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create wallet asset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Asset credited successfully",
	})
}

func (h *WalletHandler) HandleGetWallet(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	wallet, err := h.WalletService.GetWalletByUserID(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to get wallet assets", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get wallet assets"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}
