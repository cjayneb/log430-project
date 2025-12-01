package ports

import (
	"brokerx/matching-service/models"
	"context"
)

type EventProducer interface {
	SendEvent(ctx context.Context, topic string, eventType string, eventData models.Order, err error) error
	SendMatchingEvent(ctx context.Context, topic string, eventType string, order models.Order, ordersData []*models.ClaimedCandidate, recordsData []*models.ExecutionRecord, err error) error
}
