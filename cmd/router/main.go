package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/segmentio/kafka-go"
)

type Event struct {
	Type string `json:"type"`
}

type DLQ struct {
	Reason string `json:"reason"`
	Raw    string `json:"raw"`
}

func main() {
	broker := env("KAFKA_BROKER", "localhost:9092")
	in := env("IN", "esb.inbound")
	outOrders := env("OUT_ORDERS", "esb.orders")
	outPayments := env("OUT_PAYMENTS", "esb.payments")
	dlq := env("DLQ", "esb.dlq")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		Topic:          in,
		GroupID:        env("GROUP", "router-v1"),
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 0,
		MinBytes:       1,
		MaxBytes:       10e6,
	})
	defer reader.Close()

	writer := func(topic string) *kafka.Writer {
		return &kafka.Writer{Addr: kafka.TCP(broker), Topic: topic, Balancer: &kafka.Hash{}}
	}
	wOrders := writer(outOrders)
	wPayments := writer(outPayments)
	wDLQ := writer(dlq)
	defer wOrders.Close(); defer wPayments.Close(); defer wDLQ.Close()

	log.Printf("router: %s -> [%s, %s] dlq=%s\n", in, outOrders, outPayments, dlq)
	for {
		m, err := reader.FetchMessage(context.Background())
		if err != nil { log.Fatalf("fetch: %v", err) }

		var e Event
		msg := string(m.Value)
		if err := json.Unmarshal(m.Value, &e); err != nil || !validType(e.Type) {
			// send to DLQ with reason
			dq, _ := json.Marshal(DLQ{Reason: reason(err, e.Type), Raw: compact(msg)})
			_ = wDLQ.WriteMessages(context.Background(), kafka.Message{Key: []byte("dlq"), Value: dq})
			_ = reader.CommitMessages(context.Background(), m)
			continue
		}

		var target *kafka.Writer
		switch e.Type {
		case "order":
			target = wOrders
		case "payment":
			target = wPayments
		default:
			target = wDLQ
		}
		if err := target.WriteMessages(context.Background(), kafka.Message{Key: m.Key, Value: m.Value}); err != nil {
			log.Printf("route write error: %v", err)
		}
		_ = reader.CommitMessages(context.Background(), m)
	}
}

func validType(t string) bool { return t == "order" || t == "payment" }
func reason(err error, t string) string {
	if err != nil { return "invalid_json" }
	if !validType(t) { return "unknown_type" }
	return "unknown"
}
func compact(s string) string { return strings.TrimSpace(bufio.NewScanner(strings.NewReader(s)).Text()) }

func env(k, def string) string { v := os.Getenv(k); if v == "" { return def }; return v }
