package ports

import (
	"brokerx/portfolio-service/models"
	"context"
)

type EventProducer interface {
	SendEvent(ctx context.Context, topic string, eventType string, eventData models.Order, err error) error
}
