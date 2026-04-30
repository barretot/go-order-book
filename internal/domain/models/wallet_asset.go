package models

import "github.com/google/uuid"

type Instrument string

const (
	BTC  Instrument = "BTC"
	BRL  Instrument = "BRL"
	USDT Instrument = "USDT"
	ETH  Instrument = "ETH"
)

var AcceptedInstruments = []Instrument{BTC, BRL, USDT, ETH}

type WalletAssets struct {
	UserId     uuid.UUID
	Instrument Instrument
	Quantity   float64
}

type WalletAssetDBModel struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	Instrument Instrument `json:"instrument"`
	Quantity   float64    `json:"quantity"`
}

type WalletDBModel struct {
	UserID uuid.UUID            `json:"user_id"`
	Email  string               `json:"email"`
	Assets []WalletAssetDBModel `json:"assets"`
}
