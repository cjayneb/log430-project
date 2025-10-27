package models

import (
	"database/sql"
)

type Order struct {
	ID                int	  `json:"order_id"`
	UserID            int  `json:"user_id" validate:"required"`
	Symbol            string  `json:"symbol" validate:"required"`
	Type              string  `json:"type" validate:"required,oneof=market limit"`
	Action            string  `json:"action" validate:"required,oneof=buy sell"`
	Quantity          int     `json:"quantity" validate:"required,gt=0"`
	RemainingQuantity int     `json:"remaining_quantity" validate:"required,eqfield=Quantity"`
	UnitPrice         float64 `json:"unit_price" validate:"gte=0,limitprice"`
	Timing            string  `json:"timing" validate:"required,oneof=day ioc"`
	Status            string  // open, partially filled, filled, canceled
	CreatedAt         sql.NullTime
	UpdatedAt         sql.NullTime
}
