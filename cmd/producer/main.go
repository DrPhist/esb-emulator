package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/you/esb-emulator/internal/generator"
)

func main() {
	broker := env("KAFKA_BROKER", "localhost:9092")
	topic := env("TOPIC", "esb.inbound")
	rate := dur("RATE", 250*time.Millisecond) // event every 250ms

	w := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.Hash{}, // stable key routing
		BatchTimeout: 10 * time.Millisecond,
	}
	defer w.Close()

	log.Printf("producer -> %s (broker %s)\n", topic, broker)
	for {
		ev := generator.NewEvent()
		b, _ := json.Marshal(ev)
		msg := kafka.Message{Key: []byte(ev.Type), Value: b}
		if err := w.WriteMessages(context.Background(), msg); err != nil {
			log.Printf("write error: %v", err)
		}
		time.Sleep(rate)
	}
}

func env(k, def string) string { v := os.Getenv(k); if v == "" { return def }; return v }
func dur(k string, def time.Duration) time.Duration { if v := os.Getenv(k); v != "" { if d, err := time.ParseDuration(v); err == nil { return d } }; return def }
