package models

type OrderEvent struct {
	Topic   string
	Event   string
	TraceID string
	UserId  string
	JWT     string
	Order   Order
	Error   string
}
