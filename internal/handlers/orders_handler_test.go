package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrdersHandlerPlaceOrderReturnsAcceptedSidesWhenSideIsInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &OrdersHandler{}
	router := gin.New()
	router.POST("/orders/:userId", handler.HandlePlaceOrder)

	body := `{"instrument":"BTC/BRL","quantity":1,"side":"hold","price":500000}`
	request := httptest.NewRequest(http.MethodPost, "/orders/"+uuid.NewString(), strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	assert.Equal(t, "denied", payload["status"])
	assert.Equal(t, []any{"buy", "sell"}, payload["accepted_sides"])
	assert.Contains(t, payload["reason"], "Side")
}
