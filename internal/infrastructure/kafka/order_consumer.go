package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"parser/pkg/worker_pool"

	"github.com/segmentio/kafka-go"
)

type OrderRequestHandler interface {
	ProcessOrder(ctx context.Context, messageID, topic string, request OrderRequest) error
}

type OrderConsumer struct {
	reader  *kafka.Reader
	handler OrderRequestHandler
	workers *worker_pool.WorkerPool
}

func NewOrderConsumer(brokers []string, topic, groupID string, handler OrderRequestHandler, workers *worker_pool.WorkerPool) *OrderConsumer {
	return &OrderConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		handler: handler,
		workers: workers,
	}
}

func (c *OrderConsumer) Run(ctx context.Context) {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("failed to fetch kafka order request: %v", err)
			continue
		}

		if err := c.submit(ctx, message); err != nil {
			log.Printf("failed to submit kafka order request task: %v", err)
		}
	}
}

func (c *OrderConsumer) submit(ctx context.Context, message kafka.Message) error {
	if c.workers == nil {
		return c.handleMessage(ctx, message)
	}

	return c.workers.Submit(func() {
		taskCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := c.handleMessage(taskCtx, message); err != nil {
			log.Printf("failed to handle kafka order request: %v", err)
		}
	})
}

func (c *OrderConsumer) handleMessage(ctx context.Context, message kafka.Message) error {
	var request OrderRequest
	if err := json.Unmarshal(message.Value, &request); err != nil {
		return err
	}

	if err := c.handler.ProcessOrder(ctx, orderRequestMessageID(request, message), message.Topic, request); err != nil {
		return err
	}

	return c.reader.CommitMessages(ctx, message)
}

func orderRequestMessageID(request OrderRequest, message kafka.Message) string {
	if request.EventID != 0 {
		return fmt.Sprintf("%d", request.EventID)
	}
	if len(message.Key) > 0 {
		return string(message.Key)
	}
	return fmt.Sprintf("%s:%d:%d", message.Topic, message.Partition, message.Offset)
}

func (c *OrderConsumer) Close() error {
	return c.reader.Close()
}
