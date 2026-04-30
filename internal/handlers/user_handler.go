package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/barretot/go-order-book/internal/domain/dto"
	"github.com/barretot/go-order-book/internal/domain/models"
	"github.com/barretot/go-order-book/internal/services"
	"github.com/barretot/go-order-book/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	UserService *services.UserService
}

func (h *UserHandler) HandleCreateUser(c *gin.Context) {
	var request dto.CreateUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			errors := make(map[string]string)
			for _, fieldErr := range verrs {
				errors[fieldErr.Field()] = utils.GetErrorMessage(fieldErr)
			}
			slog.Error("validation request error", "error", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"status": "denied", "reason": errors})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid payload"})
		return
	}

	input := models.User{
		Name:  request.Name,
		Email: request.Email,
	}

	userID, err := h.UserService.CreateUser(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, services.ErrDuplicatedEmail) {
			slog.Warn("duplicate email detected")
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		slog.Error("failed to create user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": userID,
	})
}
