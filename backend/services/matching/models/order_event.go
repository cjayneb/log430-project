package models

type OrderEvent struct {
	Event   string
	TraceID string
	UserId  string
	JWT     string
	Order   Order
	Error   string
}
