package kafka

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type OrderProducer struct {
	writer *kafka.Writer
}

func NewOrderProducer(brokers []string, topic string) *OrderProducer {
	return &OrderProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
	}
}

func (p *OrderProducer) Publish(ctx context.Context, response OrderResponse) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(response.OrderID, 10)),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(response.EventID)},
			{Key: "event_type", Value: []byte("order.parsing_status_changed")},
		},
	})
}

func (p *OrderProducer) Close() error {
	return p.writer.Close()
}
