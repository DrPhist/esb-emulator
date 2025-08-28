package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/DrPhist/esb-emulator/internal/kafkautil"
	"github.com/segmentio/kafka-go"
)

// Event represents a Kafka message event.
type Event struct {
	Type string `json:"type"`
}

// DLQ represents a dead-letter queue message.
type DLQ struct {
	Reason string `json:"reason"`
	Raw    string `json:"raw"`
}

func main() {
	broker := env("KAFKA_BROKER", "localhost:9094")

	in := env("IN", "esb.inbound")
	outOrders := env("OUT_ORDERS", "esb.orders")
	outPayments := env("OUT_PAYMENTS", "esb.payments")
	dlq := env("DLQ", "esb.dlq")

	// Ensure topics exist before consuming/producing
	kafkautil.MustEnsureTopic(broker, kafka.TopicConfig{Topic: in, NumPartitions: 3, ReplicationFactor: 1})
	kafkautil.MustEnsureTopic(broker, kafka.TopicConfig{Topic: outOrders, NumPartitions: 3, ReplicationFactor: 1})
	kafkautil.MustEnsureTopic(broker, kafka.TopicConfig{Topic: outPayments, NumPartitions: 3, ReplicationFactor: 1})
	kafkautil.MustEnsureTopic(broker, kafka.TopicConfig{Topic: dlq, NumPartitions: 1, ReplicationFactor: 1})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		Topic:          in,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 0,
		MinBytes:       1,
		MaxBytes:       10e6,
	})

	defer reader.Close()

	mkWriter := func(topic string) *kafka.Writer {
		return &kafka.Writer{Addr: kafka.TCP(broker), Topic: topic, Balancer: &kafka.Hash{}}
	}

	wOrders := mkWriter(outOrders)
	wPayments := mkWriter(outPayments)
	wDLQ := mkWriter(dlq)

	defer wOrders.Close()
	defer wPayments.Close()
	defer wDLQ.Close()

	log.Printf("router: %s -> [%s, %s] dlq=%s", in, outOrders, outPayments, dlq)
	for {
		m, err := reader.FetchMessage(context.Background())
		if err != nil {
			log.Fatalf("fetch: %v", err)
		}

		var e Event

		msg := string(m.Value)
		if err := json.Unmarshal(m.Value, &e); err != nil || !validType(e.Type) {
			dq, _ := json.Marshal(DLQ{Reason: reason(err, e.Type), Raw: compact(msg)})
			if err := wDLQ.WriteMessages(context.Background(), kafka.Message{Key: []byte("dlq"), Value: dq}); err != nil {
				log.Printf("dlq write error: %v", err)
			}
			_ = reader.CommitMessages(context.Background(), m)
			continue
		}

		target := map[string]*kafka.Writer{"order": wOrders, "payment": wPayments}[e.Type]
		if target == nil {
			target = wDLQ
		}

		if err := target.WriteMessages(context.Background(), kafka.Message{Key: m.Key, Value: m.Value}); err != nil {
			log.Printf("route write error: %v", err)
		}

		_ = reader.CommitMessages(context.Background(), m)
	}
}

// validType checks if the event type is valid.
func validType(t string) bool { return t == "order" || t == "payment" }

// reason returns the reason for a dead-letter queue message.
func reason(err error, t string) string {
	if err != nil {
		return "invalid_json"
	}

	if !validType(t) {
		return "unknown_type"
	}

	return "unknown"
}

// compact removes all whitespace from the beginning and end of a string.
func compact(s string) string {
	sc := bufio.NewScanner(strings.NewReader(s))
	if sc.Scan() {
		return strings.TrimSpace(sc.Text())
	}

	return strings.TrimSpace(s)
}

// env retrieves an environment variable or returns a default value.
func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}

	return def
}
