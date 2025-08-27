package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"order/internal/model"

	"github.com/segmentio/kafka-go"
)

// Consumer consumes messages from a Kafka topic and validates orders.
type Consumer struct {
	reader    *kafka.Reader
	dlqWriter *kafka.Writer
}

// NewConsumer creates a new Kafka consumer with the given broker, topic, group ID, and DLQ topic.
func NewConsumer(broker, topic, groupID, dlqTopic string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID,
	})

	dlqWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   dlqTopic,
	})

	return &Consumer{reader: reader, dlqWriter: dlqWriter}
}

// Start begins consuming messages from Kafka and processes valid orders using the handle function.
// Invalid messages or messages failing validation are sent to the DLQ.
func (c *Consumer) Start(ctx context.Context, handle func(orders []model.Order) error) error {
	slog.Info("Kafka consumer started")
	defer c.reader.Close()
	defer c.dlqWriter.Close()

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
			c.sendToDLQ(ctx, m)
			continue
		}

		validOrders := make([]model.Order, 0, len(orders))
		for _, order := range orders {
			if err := ValidateOrder(&order); err != nil {
				slog.Warn("invalid order", slog.String("order_uid", order.OrderUID), slog.String("error", err.Error()))
				c.sendToDLQ(ctx, kafka.Message{
					Key:   []byte(order.OrderUID),
					Value: m.Value,
				})
				continue
			}
			validOrders = append(validOrders, order)
		}

		if len(validOrders) == 0 {
			continue
		}

		if err := handle(validOrders); err != nil {
			slog.Error("failed to handle orders", slog.String("error", err.Error()))
			c.sendToDLQ(ctx, m)
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			slog.Error("failed to commit message", slog.String("error", err.Error()))
		}
	}
}

// sendToDLQ sends a Kafka message to the Dead Letter Queue.
func (c *Consumer) sendToDLQ(ctx context.Context, msg kafka.Message) {
	if err := c.dlqWriter.WriteMessages(ctx, msg); err != nil {
		slog.Error("failed to write message to DLQ", slog.String("error", err.Error()))
	}
}

// Close closes the Kafka consumer reader and DLQ writer.
func (c *Consumer) Close() error {
	c.dlqWriter.Close()
	return c.reader.Close()
}
