package kafkautil

import (
	"log"

	"github.com/segmentio/kafka-go"
)

// EnsureTopic creates a topic with the given configuration if it does not exist.
// The call is idempotent on modern Kafka brokers.
func EnsureTopic(broker string, cfg kafka.TopicConfig) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}

	defer conn.Close()

	if err := conn.CreateTopics(cfg); err != nil {
		// Some brokers return TopicAlreadyExists as an error; log and continue.
		log.Printf("EnsureTopic: %s -> %v", cfg.Topic, err)
		return err
	}

	return nil
}

// MustEnsureTopic wraps EnsureTopic and only logs errors.
func MustEnsureTopic(broker string, cfg kafka.TopicConfig) {
	if err := EnsureTopic(broker, cfg); err != nil {
		log.Printf("MustEnsureTopic warn: %v", err)
	}
}
