package models

type Instrument struct {
	Symbol           string  `json:"Symbol"`
	Name             string  `json:"Name"`
	TickSize         float64 `json:"TickSize"`
	PriceBandPercent int     `json:"PriceBandPercent"`
	Status           string  `json:"Status"`
}