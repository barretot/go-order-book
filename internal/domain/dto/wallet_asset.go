package dto

type AddWalletAssetRequest struct {
	Instrument string  `json:"instrument" binding:"required,oneof=BTC BRL USDT ETH"`
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
}
