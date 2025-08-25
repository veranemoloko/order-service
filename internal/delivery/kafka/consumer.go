package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	model "order/internal/entity"

	"github.com/segmentio/kafka-go"
)

// Consumer consumes messages from a Kafka topic and validates orders.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a new Kafka consumer with the given broker, topic, and group ID.
func NewConsumer(broker, topic, groupID string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID,
	})

	return &Consumer{reader: r}
}

// Start begins consuming messages from Kafka and processes valid orders using the handle function.
func (c *Consumer) Start(ctx context.Context, handle func(orders []model.Order) error) error {
	slog.Info("Kafka consumer started")
	defer c.reader.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigchan
		slog.Info("Stopping Kafka consumer...")
		cancel()
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("error reading message", slog.String("error", err.Error()))
			continue
		}

		var orders []model.Order
		if err := json.Unmarshal(m.Value, &orders); err != nil || len(orders) == 0 {
			slog.Error("failed to unmarshal order", slog.String("error", err.Error()))
		}

		validOrders := make([]model.Order, 0, len(orders))
		for _, order := range orders {
			if err := ValidateOrder(&order); err != nil {
				slog.Warn("invalid order", slog.String("order_uid", order.OrderUID), slog.String("error", err.Error()))
				continue
			}
			validOrders = append(validOrders, order)
		}

		if err := handle(validOrders); err != nil {
			slog.Error("failed to handle orders", slog.String("error", err.Error()))
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			slog.Error("failed to commit message", slog.String("error", err.Error()))
		}
	}
}

// Close closes the Kafka consumer reader.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
