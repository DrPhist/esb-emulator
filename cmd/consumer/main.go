package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	topic := os.Getenv("TOPIC")
	if topic == "" {
		topic = "esb.orders" // change to esb.payments or esb.dlq
	}

	group := os.Getenv("GROUP")
	if group == "" {
		group = "consumer-v1"
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		Topic:          topic,
		GroupID:        group,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: time.Second,
	})

	defer r.Close()

	log.Printf("consumer -> %s (broker %s)\n", topic, broker)
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("read: %v", err)
		}
		fmt.Printf("[%s] key=%s value=%s\n", topic, string(m.Key), string(m.Value))
	}
}
