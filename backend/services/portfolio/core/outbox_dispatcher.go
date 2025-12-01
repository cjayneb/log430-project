package core

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"log/slog"
	"time"
)

type OutboxDispatcher struct {
	Repo         ports.OutboxRepository
	Producer     ports.EventProducer
	PollInterval time.Duration
	BatchSize    int
	MaxRetries   int
	BaseBackoff  time.Duration
	StopChan     chan struct{}
}

func NewOutboxDispatcher(repo ports.OutboxRepository, producer ports.EventProducer, pollInterval time.Duration, batchSize int, maxRetries int, baseBackoff time.Duration) *OutboxDispatcher {
	return &OutboxDispatcher{
		Repo:         repo,
		Producer:     producer,
		PollInterval: pollInterval,
		BatchSize:    batchSize,
		MaxRetries:   maxRetries,
		BaseBackoff:  baseBackoff,
		StopChan:     make(chan struct{}),
	}
}

func (d *OutboxDispatcher) Start() {
	slog.Info("Starting outbox dispatcher!")
	go d.loop()
}

func (d *OutboxDispatcher) Stop() {
	slog.Info("Stopping outbox dispatcher!")
	close(d.StopChan)
}

func (d *OutboxDispatcher) loop() {
	ticker := time.NewTicker(d.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.processOnce(context.Background())

		case <-d.StopChan:
			return
		}
	}
}

func (d *OutboxDispatcher) processOnce(ctx context.Context) {
	log := util.FromContext(ctx)

	events, err := d.Repo.FetchPending(ctx, d.BatchSize)
	if err != nil {
		log.Error("outbox fetch failed", "error", err)
		return
	}

	for _, rec := range events {
		ctx = context.WithValue(ctx, util.CtxKeyTraceId, rec.TraceID)
		ctx = context.WithValue(ctx, util.CtxKeyUserId, rec.UserID)
		ctx = context.WithValue(ctx, util.CtxKeyJWT, rec.JWT)
		ctx = util.WithLogger(ctx, log)
		if err := d.processRecord(ctx, rec); err != nil {
			log.Error("failed to process outbox record", "id", rec.ID, "error", err)
		}
	}
}

func (d *OutboxDispatcher) processRecord(ctx context.Context, rec models.OutboxRecord) error {
	log := util.FromContext(ctx)

	var order models.Order
	if err := rec.UnmarshalPayload(&order); err != nil {
		log.Error("payload unmarshal failed", "id", rec.ID, "error", err)
		return d.Repo.MarkFailed(ctx, rec.ID, "invalid payload")
	}

	err := d.Producer.SendEvent(ctx, rec.Topic, rec.EventType, order, nil)
	if err == nil {
		return d.Repo.MarkPublished(ctx, rec.ID)
	}

	log.Error("send failed", "id", rec.ID, "error", err)
	if rec.RetryCount+1 >= d.MaxRetries {
		return d.Repo.MarkFailed(ctx, rec.ID, err.Error())
	}

	// schedule next retry using backoff: base * 2^retry
	next := time.Now().Add(d.BaseBackoff * time.Duration(1<<rec.RetryCount))
	return d.Repo.IncrementRetry(ctx, rec.ID, next)
}
