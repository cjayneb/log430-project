package dao_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type SQLOutboxRepository struct {
	DB *sql.DB
	tx *sql.Tx
}

func NewOutboxRepo(tx *sql.Tx) SQLOutboxRepository {
	return SQLOutboxRepository{tx: tx}
}

func (s SQLOutboxRepository) FetchPending(ctx context.Context, limit int) ([]models.OutboxRecord, error) {
	log := util.FromContext(ctx)

	stmt := `
        SELECT id, topic, event_type, trace_id, user_id, jwt_token, payload, retry_count
        FROM outbox_order_events
        WHERE status='pending' AND next_attempt_at <= NOW()
        ORDER BY id
        LIMIT ?
        FOR UPDATE SKIP LOCKED
    `

	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Error("error beginning transaction", "error", err)
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, stmt, limit)
	if err != nil {
		log.Error("error executing query", "error", err)
		_ = tx.Rollback()
		return nil, err
	}
	defer rows.Close()

	records := []models.OutboxRecord{}
	for rows.Next() {
		var rec models.OutboxRecord
		err = rows.Scan(&rec.ID, &rec.Topic, &rec.EventType, &rec.TraceID, &rec.UserID, &rec.JWT, &rec.PayloadString, &rec.RetryCount)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		records = append(records, rec)
	}

	return records, tx.Commit()
}

func (s SQLOutboxRepository) IncrementRetry(ctx context.Context, id int64, next time.Time) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE outbox_order_events 
            SET retry_count = retry_count + 1,
                next_attempt_at = ?,
                updated_at=NOW()
         WHERE id=?`,
		next, id,
	)
	return err
}

func (s SQLOutboxRepository) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE outbox_order_events SET status='failed', error_message=? WHERE id=?`,
		errMsg, id,
	)
	return err
}

func (s SQLOutboxRepository) MarkPublished(ctx context.Context, id int64) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE outbox_order_events SET status='published' WHERE id=?`,
		id,
	)
	return err
}

func (s SQLOutboxRepository) CreateOrderEvents(ctx context.Context, events []*models.OrderEvent) error {
	log := util.FromContext(ctx)

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
