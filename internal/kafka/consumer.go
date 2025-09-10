package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	kfk "github.com/segmentio/kafka-go"
	"github.com/Uva337/WBL0v1/internal/interfaces"
	"github.com/Uva337/WBL0v1/internal/models"
)

type Handler func(ctx context.Context, o models.Order) error

type Consumer struct {
	r         *kfk.Reader
	validator interfaces.Validator
}

func NewConsumer(v interfaces.Validator) *Consumer {
	brokers := envOr("KAFKA_BROKERS", "localhost:9092")
	topic := envOr("KAFKA_TOPIC", "orders")
	group := envOr("KAFKA_GROUP", "order-demo-consumer")

	r := kfk.NewReader(kfk.ReaderConfig{
		Brokers:           []string{brokers},
		GroupID:           group,
		Topic:             topic,
		MinBytes:          1,
		MaxBytes:          10e6,
		StartOffset:       kfk.LastOffset,
		HeartbeatInterval: 3 * time.Second,
		CommitInterval:    time.Second,
	})
	return &Consumer{r: r, validator: v}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func (c *Consumer) Close() {
	if err := c.r.Close(); err != nil {
		log.Printf("failed to close kafka reader: %v", err)
	}
}

func (c *Consumer) Run(ctx context.Context, h Handler) error {
	for {
		m, err := c.r.ReadMessage(ctx)
		if err != nil {
			
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		var o models.Order
		if err := json.Unmarshal(m.Value, &o); err != nil {
			log.Printf("skip invalid json message: %v", err)
			continue
		}

		
		if err := c.validator.Struct(o); err != nil {
			log.Printf("skip invalid order data (uid: %s): %v", o.OrderUID, err)
			continue
		}

		if err := h(ctx, o); err != nil {
			log.Printf("handler error for order %s: %v", o.OrderUID, err)
		}
	}
}
