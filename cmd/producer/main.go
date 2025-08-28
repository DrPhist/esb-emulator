package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/DrPhist/esb-emulator/internal/generator"
	"github.com/DrPhist/esb-emulator/internal/kafkautil"
	"github.com/segmentio/kafka-go"
)

func main() {

	// Use the host-facing listener exposed in docker-compose.
	// Override with KAFKA_BROKER if needed.
	broker := env("KAFKA_BROKER", "localhost:9094")
	topic := env("TOPIC", "esb.inbound")
	rate := dur("RATE", 250*time.Millisecond) // one event every 250ms

	// Ensure the inbound topic exists before producing.
	kafkautil.MustEnsureTopic(broker, kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     3,
		ReplicationFactor: 1,
	})

	w := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		BatchTimeout: 10 * time.Millisecond,
	}

	defer w.Close()

	log.Printf("producer -> %s (broker %s)\n", topic, broker)

	for {
		ev := generator.NewEvent()
		b, _ := json.Marshal(ev)

		// Key by event type; you could also key by order_id/payment_id for per-entity ordering.
		msg := kafka.Message{Key: []byte(ev.Type), Value: b}

		if err := w.WriteMessages(context.Background(), msg); err != nil {
			log.Printf("write error: %v", err)
		}

		time.Sleep(rate)
	}
}

// env retrieves an environment variable or returns a default value.
func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}

	return def
}

// dur retrieves a duration from an environment variable or returns a default value.
func dur(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}

	return def
}
