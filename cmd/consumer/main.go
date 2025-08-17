package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/DrPhist/esb-emulator/internal/kafkautil"
	"github.com/segmentio/kafka-go"
)

func main() {
	broker := env("KAFKA_BROKER", "localhost:9094")
	topic := env("TOPIC", "esb.orders")
	group := env("GROUP", "consumer-v1")

	// Ensure the topic exists. Use 1 partition for DLQ by convention.
	parts := 3
	if topic == "esb.dlq" {
		parts = 1
	}

	kafkautil.MustEnsureTopic(broker, kafka.TopicConfig{Topic: topic, NumPartitions: parts, ReplicationFactor: 1})

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		Topic:          topic,
		GroupID:        group,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: time.Second,
	})

	defer r.Close()

	log.Printf("consumer -> %s (broker %s)", topic, broker)
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("read: %v", err)
		}
		fmt.Printf("[%s] key=%s value=%s\n", topic, string(m.Key), string(m.Value))
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}

	return def
}
