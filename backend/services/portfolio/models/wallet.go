package models

type Wallet struct {
	ID             string
	UserId         int
	AvailableFunds float64 `json:"available_funds"`
	ReservedFunds  float64 `json:"reserved_funds"`
}