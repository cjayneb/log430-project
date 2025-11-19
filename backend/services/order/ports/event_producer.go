package ports

import "context"

type EventProducer interface {
	SendEvent(ctx context.Context, topic string, eventData any) error
}
