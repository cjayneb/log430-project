package models

import "database/sql"

type Order struct {
	ID        int
	UserID    string `schema:"user_id"`
	Symbol	  string `schema:"symbol"`
	Type      string `schema:"type"`  // market, limit
	Action	  string `schema:"action"`  // buy, sell
	Quantity  int `schema:"quantity"`
	UnitPrice float64 `schema:"unit_price"`
	Timing	  string  `schema:"timing"` // day, ioc 	
	Status	  string `schema:"status"` // open, partially filled, filled, canceled
	CreatedAt  sql.NullTime
	UpdatedAt  sql.NullTime 
}