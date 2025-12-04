package ports

import (
	"brokerx/notification-service/models"
	"context"
)

type EventProducer interface {
	SendEvent(ctx context.Context, event models.OrderEvent) error
}
