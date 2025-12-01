package dao_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"database/sql"
	"encoding/json"
)

type SQLOutboxRepository struct {
	DB *sql.DB
	tx *sql.Tx
}

func NewOutboxRepo(tx *sql.Tx) SQLOutboxRepository {
	return SQLOutboxRepository{tx: tx}
}

func (s SQLOutboxRepository) CreateOrderEvents(ctx context.Context, events []*models.OrderEvent) error {
	log := common.FromContext(ctx)

	if len(events) == 0 {
		return nil
	}

	stmt := `
		INSERT INTO outbox_order_events
			(topic, event_type, trace_id, user_id, jwt_token, payload)
		VALUES
			(?, ?, ?, ?, ?, ?)
	`

	for _, ev := range events {
		payloadBytes, err := json.Marshal(ev.Order)
		if err != nil {
			log.Error("failed to marshal outbox payload", "error", err)
			return err
		}

		_, err = s.tx.ExecContext(
			ctx,
			stmt,
			ev.Topic,
			ev.Event,
			ev.TraceID,
			ev.UserId,
			ev.JWT,
			string(payloadBytes),
		)

		if err != nil {
			log.Error("failed to insert outbox event", "event", ev.Event, "error", err)
			return err
		}
	}

	log.Info("successfully created outbox events", "count", len(events))
	return nil
}

var _ ports.OutboxRepository = (*SQLOutboxRepository)(nil) // Ensure interface is implemented at compile time
