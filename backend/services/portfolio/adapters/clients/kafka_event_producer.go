package client_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type MatchingEvent struct {
	Event      string
	TraceID    string
	UserId     string
	JWT        string
	Order      models.Order
	Orders     []*models.Order
	Executions []*models.ExecutionRecord
	Error      string
}

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
	log := util.FromContext(ctx)

	traceId := ctx.Value(util.CtxKeyTraceId)
	if traceId == nil {
		msg := "missing traceId. cannot send event"
		log.Error(msg)
		return errors.New(msg)
	}
	userId := ctx.Value(util.CtxKeyUserId)
	if userId == nil {
		userId = ""
	}
	jwt := ctx.Value(util.CtxKeyJWT)
	if jwt == nil {
		jwt = ""
	}
	
	errString := ""
	if err != nil {
		errString = err.Error()
	}

	event := models.OrderEvent{
		Event:   eventType,
		TraceID: traceId.(string),
		UserId:  userId.(string),
		JWT:     jwt.(string),
		Order:   eventData,
		Error:   errString,
	}

	jsonEventData, err := json.Marshal(event)
	if err != nil {
		log.Error("error when encoding JSON", "error", err)
		return err
	}

	err = k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          jsonEventData,
		Headers:        []kafka.Header{{Key: string(util.HeaderTraceId), Value: []byte(traceId.(string))}},
	}, nil)
	if err != nil {
		log.Error("error when producing event", "error", err)
		return err
	}

	return nil
}

var _ ports.EventProducer = (*KafkaEventProducer)(nil) // Ensure interface is implemented at compile time
