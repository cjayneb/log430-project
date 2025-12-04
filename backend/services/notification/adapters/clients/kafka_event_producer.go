package client_adapters

import (
	"brokerx/notification-service/models"
	"brokerx/notification-service/ports"
	"brokerx/notification-service/util"
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

func (k *KafkaEventProducer) SendEvent(ctx context.Context, event models.OrderEvent) error {
	log := util.FromContext(ctx)

	log.Info("event to send", "event", event)

	jsonEventData, err := json.Marshal(event)
	if err != nil {
		log.Error("error when encoding JSON", "error", err)
		return err
	}

	err = k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &event.Topic, Partition: kafka.PartitionAny},
		Value:          jsonEventData,
		Headers:        []kafka.Header{{Key: string(util.HeaderTraceId), Value: []byte(event.TraceID)}},
	}, nil)
	if err != nil {
		log.Error("error when producing event", "error", err)
		return err
	}

	return nil
}

var _ ports.EventProducer = (*KafkaEventProducer)(nil) // Ensure interface is implemented at compile time
