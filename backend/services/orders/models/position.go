package models

type Position struct {
	ID        int
	UserId    string
	Symbol    string
	Quantity  int
	UnitPrice float64
}