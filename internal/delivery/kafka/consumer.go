package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"

	model "order/internal/entity"
)

// Consumer consumes messages from a Kafka topic and validates orders.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a new Kafka consumer with the given broker, topic, and group ID.
func NewConsumer(broker, topic, groupID string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{reader: r}
}

// Start begins consuming messages from Kafka and processes valid orders using the handle function.
func (c *Consumer) Start(ctx context.Context, handle func(orders []model.Order)) error {
	log.Println("Kafka consumer started...")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigchan
		log.Println("Stopping Kafka consumer...")
		cancel()
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("error reading message: %v", err)
			continue
		}

		var orders []model.Order

		// Try to parse as an array first
		if err := json.Unmarshal(m.Value, &orders); err != nil || len(orders) == 0 {
			// If not an array, try parsing as a single order
			var order model.Order
			if err := json.Unmarshal(m.Value, &order); err != nil {
				log.Printf("failed to unmarshal order: %v", err)
				continue
			}
			orders = append(orders, order)
		}

		validOrders := make([]model.Order, 0, len(orders))
		for _, order := range orders {
			if err := ValidateOrder(&order); err != nil {
				log.Printf("invalid order %s: %v", order.OrderUID, err)
				continue
			}
			validOrders = append(validOrders, order)
		}

		if len(validOrders) > 0 {
			handle(validOrders)
		} else {
			log.Println("No valid orders found")
		}
	}
}

// Close closes the Kafka consumer reader.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
