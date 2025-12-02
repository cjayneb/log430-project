package models

type Position struct {
	ID        int `json:"id"`
	UserId    int
	Symbol    string `json:"symbol"`
	AvailableQuantity  int `json:"available_quantity"`
	ReservedQuantity int `json:"reserved_quantity"`
	UnitPrice float64
}