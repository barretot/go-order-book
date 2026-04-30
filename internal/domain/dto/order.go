package dto

type PlaceOrderRequest struct {
	Instrument string  `json:"instrument" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required,min=1"`
	Side       string  `json:"side" binding:"required,oneof=buy sell"`
	Price      float64 `json:"price" binding:"required,min=1"`
}
