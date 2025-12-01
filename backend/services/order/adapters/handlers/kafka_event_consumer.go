package handler_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaEventConsumer struct {
	consumer                     *kafka.Consumer
	orderCreatedHandler          OrderCreatedHandler
	orderFailedHandler OrderFailedHandler
}

func NewKafkaEventConsumer(host string, groupId string, orderCreatedHandler OrderCreatedHandler, orderFailedHandler OrderFailedHandler) *KafkaEventConsumer {
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

	return &KafkaEventConsumer{
		consumer:                     c,
		orderCreatedHandler:          orderCreatedHandler,
		orderFailedHandler: orderFailedHandler,
	}
}

func (k *KafkaEventConsumer) Start(topic string) {
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

func (k *KafkaEventConsumer) handleMessage(msg *kafka.Message) {
	var log *slog.Logger
	for _, header := range msg.Headers {
		if header.Key == string(common.HeaderTraceId) {
			log = common.FromEvent(string(header.Value))
		}
	}

	var event models.OrderEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Error("unable to  unmarshal event", "error", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.CtxKeyTraceId, event.TraceID)
	ctx = context.WithValue(ctx, common.CtxKeyUserId, event.UserId)
	ctx = context.WithValue(ctx, common.CtxKeyJWT, event.JWT)
	ctx = common.WithLogger(ctx, log)

	switch event.Event {
	case "OrderCreated":
		log.Info("Received OrderCreated event.")
		err := k.orderCreatedHandler.handle(ctx, event)
		if err != nil {
			log.Error("error handling OrderCreated event", "error", err)
			break
		}
		k.consumer.CommitMessage(msg)
	case "OrderCreatedFailed", "OrderComplianceFailed", "OrderQueuingFailed", "OrderMatchingFailed", "OrderConfirmationFailed":
		log.Info("Received OrderFailed event", "event", event.Event)
		err := k.orderFailedHandler.handle(ctx, event)
		if err != nil {
			log.Error("error handling OrderFailed event", "error", err)
			break
		}
		k.consumer.CommitMessage(msg)
	case "OrderSagaCompleted":
		if event.Error != "" {
			log.Error("Order Saga completed with an error", "error", event.Error)
		} else {
			log.Info("Order Saga completed!", "orderId", event.Order.ID)
		}
	}
}

var _ ports.EventConsumer = (*KafkaEventConsumer)(nil) // Ensure interface is implemented at compile time
