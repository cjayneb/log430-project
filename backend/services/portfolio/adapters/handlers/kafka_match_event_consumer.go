package handler_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaMatchingEventConsumer struct {
	consumer            *kafka.Consumer
	orderMatchedHandler OrderMatchedHandler
}

func NewKafkaMatchingEventConsumer(host string, groupId string, orderMatchedHandler OrderMatchedHandler) *KafkaMatchingEventConsumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  host,
		"group.id":           groupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})

	if err != nil {
		slog.Error("Could not create kafka consumer", "error", err)
		os.Exit(1)
	}

	return &KafkaMatchingEventConsumer{
		consumer:            c,
		orderMatchedHandler: orderMatchedHandler,
	}
}

func (k *KafkaMatchingEventConsumer) Start(topic string) {
	err := k.consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		slog.Error("Unable to subscribe to topic", "error", err)
		os.Exit(1)
	}
	slog.Info("Event consumer listening to topic", "topic", topic)

	for {
		event := k.consumer.Poll(50)
		switch e := event.(type) {
		case *kafka.Message:
			k.handleMessage(e)
		case kafka.Error:
			slog.Error("error when consuming event", "error", e)
		default:
			slog.Debug("Ignored", "event", e)
		}
	}
}

func (k *KafkaMatchingEventConsumer) handleMessage(msg *kafka.Message) {
	var log *slog.Logger
	for _, header := range msg.Headers {
		if header.Key == string(util.HeaderTraceId) {
			log = util.FromEvent(string(header.Value))
		}
	}

	var event models.MatchingEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Error("unable to  unmarshal event", "error", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, util.CtxKeyTraceId, event.TraceID)
	ctx = context.WithValue(ctx, util.CtxKeyUserId, event.UserId)
	ctx = context.WithValue(ctx, util.CtxKeyJWT, event.JWT)
	ctx = util.WithLogger(ctx, log)

	switch event.Event {
	case "OrderMatched":
		log.Info("Received OrderMatched event.", "orderId", event.Order.ID)
		err := k.orderMatchedHandler.handle(ctx, event)
		if err != nil {
			log.Error("error handling OrderMatched event", "error", err)
			break
		}
		_, _ = k.consumer.CommitMessage(msg)
	}
}

var _ ports.EventConsumer = (*KafkaMatchingEventConsumer)(nil) // Ensure interface is implemented at compile time
