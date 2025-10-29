package models

type Wallet struct {
	ID             string
	UserId         string
	AvailableFunds float64
	OnHoldFunds    float64
}