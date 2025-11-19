package handler_adapters

import (
	"brokerx/matching-service/ports"
	"log/slog"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaEventConsumer struct {
	consumer *kafka.Consumer
}

func NewKafkaEventConsumer(host string, groupId string) *KafkaEventConsumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": host,
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		slog.Error("Could not create kafka consumer", "error", err)
		os.Exit(1)
	}
	return &KafkaEventConsumer{consumer: c}
}

func (k *KafkaEventConsumer) Start(topic string) {
	err := k.consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		slog.Error("Unable to subscribe to topic", "error", err)
		os.Exit(1)
	}

	for {
		event := k.consumer.Poll(50)
		switch e := event.(type) {
		case *kafka.Message:
			slog.Info("Consuming event", "event", e)
			handleMessage(e)
		case kafka.Error:
			slog.Error("error when consuming event", "error", e)
		default:
			slog.Debug("Ignored", "event", e)
		}
	}
}

func handleMessage(msg *kafka.Message) {
	slog.Info("Handling message", "value", string(msg.Value))
}

var _ ports.EventConsumer = (*KafkaEventConsumer)(nil) // Ensure interface is implemented at compile time
