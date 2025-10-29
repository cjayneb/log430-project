package models

type Wallet struct {
	ID             string
	UserId         string
	AvailableFunds float64 `json:"available_funds"`
	OnHoldFunds    float64
}