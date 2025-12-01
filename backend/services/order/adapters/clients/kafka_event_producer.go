package client_adapters

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

type KafkaEventProducer struct {
	producer *kafka.Producer
}

func NewKafkaEventProducer(host string) *KafkaEventProducer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": host})
	if err != nil {
		slog.Error("Could not create kafka producer", "error", err)
		os.Exit(1)
	}
	return &KafkaEventProducer{producer: p}
}

func (k *KafkaEventProducer) SendEvent(ctx context.Context, topic string, eventType string, eventData models.Order, err error) error {
	log := common.FromContext(ctx)

	traceId := ctx.Value(common.CtxKeyTraceId).(string)
	errString := ""
	if err != nil {
		errString = err.Error()
	}

	event := models.OrderEvent{
		Event: eventType,
		TraceID: traceId,
		UserId: ctx.Value(common.CtxKeyUserId).(string),
		JWT: ctx.Value(common.CtxKeyJWT).(string),
		Order: eventData,
		Error: errString,
	}

	jsonEventData, err := json.Marshal(event)
	if err != nil {
		log.Error("error when encoding JSON", "error", err)
		return err
	}

	err = k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          jsonEventData,
		Headers: []kafka.Header{{Key: string(common.HeaderTraceId), Value: []byte(traceId)}},
	}, nil)
	if err != nil {
		log.Error("error when producing event", "error", err)
		return err
	}

	return nil
}

var _ ports.EventProducer = (*KafkaEventProducer)(nil) // Ensure interface is implemented at compile time
