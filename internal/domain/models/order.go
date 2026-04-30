package models

import "github.com/google/uuid"

type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

var AcceptedSides = []Side{Buy, Sell}

type Order struct {
	UserID     uuid.UUID
	Instrument Instrument
	Quantity   float64
	Price      float64
	Side       Side
}

type OrderDBModel struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	Instrument        Instrument `json:"instrument"`
	Quantity          float64    `json:"quantity"`
	RemainingQuantity float64    `json:"remaining_quantity"`
	Price             float64    `json:"price"`
	Side              Side       `json:"side"`
	Status            string     `json:"status"`
}

type OrderBook struct {
	Instrument Instrument     `json:"instrument"`
	Bids       []OrderDBModel `json:"bids"`
	Asks       []OrderDBModel `json:"asks"`
}
