# ESB Emulator

A Go-based Enterprise Service Bus (ESB) emulator that demonstrates event-driven architecture using Apache Kafka. This project simulates a typical ESB pattern where messages are produced, routed to different topics based on their type, and consumed by various services.

## Overview

The ESB Emulator consists of three main components:

- **Producer**: Generates synthetic events (orders, payments, etc.) and publishes them to Kafka
- **Router**: Consumes events from the inbound topic and routes them to appropriate destination topics based on event type
- **Consumer**: Subscribes to specific topics to process routed messages

## Architecture

```
[Producer] → esb.inbound → [Router] → esb.orders
                                   → esb.payments
                                   → esb.dlq (dead letter queue)
                         
[Consumer] ← esb.orders
[Consumer] ← esb.payments  
[Consumer] ← esb.dlq
```

## Components

### Docker Compose (`docker-compose.yml`)

The `docker-compose.yml` file orchestrates the entire system with the following services:

- **kafka**: Apache Kafka broker using Bitnami's KRaft-enabled image (version 3.8.0)
  - Exposes port 9092 for internal container communication
  - Exposes port 9094 for host access
  - Includes health checks to ensure proper startup
  
- **kafka-ui**: Web UI for monitoring Kafka topics, messages, and cluster health
  - Accessible at `http://localhost:8080`
  
- **init-topics**: One-shot service that creates required Kafka topics:
  - `esb.inbound` (3 partitions) - Entry point for all events
  - `esb.orders` (3 partitions) - Routed order events
  - `esb.payments` (3 partitions) - Routed payment events
  - `esb.dlq` (1 partition) - Dead letter queue for failed routing
  
- **router**: Routes messages from inbound to appropriate destination topics
- **producer**: Generates and publishes synthetic events

## Getting Started

### Prerequisites

- Docker and Docker Compose installed
- Go 1.23.4 (if running locally without Docker)

### Running the Full System

1. Start all services using Docker Compose:
```bash
docker-compose up -d
```

2. Monitor the logs:
```bash
docker-compose logs -f
```

3. Access Kafka UI at `http://localhost:8080` to view topics and messages

4. Stop the system:
```bash
docker-compose down
```

### Running the Consumer

The consumer is not included in the Docker Compose setup by default, allowing you to run it locally or in different configurations.

#### Option 1: Run Consumer Locally

1. Configure your Go environment:
```bash
export KAFKA_BROKER=localhost:9094
export TOPIC=esb.orders  # or esb.payments, esb.dlq
export GROUP=consumer-v1
```

2. Run the consumer:
```bash
go run cmd/consumer/main.go
```

#### Option 2: Run Multiple Consumers

You can run multiple consumers to process different topics simultaneously:

```bash
# Terminal 1 - Orders consumer
KAFKA_BROKER=localhost:9094 TOPIC=esb.orders GROUP=orders-consumer go run cmd/consumer/main.go

# Terminal 2 - Payments consumer  
KAFKA_BROKER=localhost:9094 TOPIC=esb.payments GROUP=payments-consumer go run cmd/consumer/main.go

# Terminal 3 - DLQ consumer
KAFKA_BROKER=localhost:9094 TOPIC=esb.dlq GROUP=dlq-consumer go run cmd/consumer/main.go
```

#### Option 3: Add Consumer to Docker Compose

You can add a consumer service to your `docker-compose.yml`:

```yaml
consumer:
  build:
    dockerfile: Dockerfile.consumer  # You'll need to create this
  container_name: consumer
  depends_on:
    kafka:
      condition: service_healthy
    init-topics:
      condition: service_completed_successfully
  environment:
    - KAFKA_BROKER=kafka:9092
    - TOPIC=esb.orders
    - GROUP=consumer-v1
```

### Configuration

All components support configuration via environment variables:

#### Producer
- `KAFKA_BROKER`: Kafka broker address (default: `localhost:9094`)
- `TOPIC`: Target topic (default: `esb.inbound`)
- `RATE`: Event generation rate (default: `250ms`)

#### Router  
- `KAFKA_BROKER`: Kafka broker address (default: `localhost:9094`)
- `IN`: Input topic (default: `esb.inbound`)
- `OUT_ORDERS`: Orders output topic (default: `esb.orders`)
- `OUT_PAYMENTS`: Payments output topic (default: `esb.payments`)
- `DLQ`: Dead letter queue topic (default: `esb.dlq`)

#### Consumer
- `KAFKA_BROKER`: Kafka broker address (default: `localhost:9094`)
- `TOPIC`: Topic to consume from (default: `esb.orders`)
- `GROUP`: Consumer group ID (default: `consumer-v1`)

## Development

### Project Structure

```
├── cmd/
│   ├── consumer/main.go    # Consumer application
│   ├── producer/main.go    # Producer application
│   └── router/main.go      # Router application
├── internal/
│   ├── generator/          # Event generation utilities
│   └── kafkautil/          # Kafka helper functions
├── docker-compose.yml      # Multi-service orchestration
├── Dockerfile.producer     # Producer container
├── Dockerfile.router       # Router container
├── go.mod                  # Go module definition
└── go.sum                  # Go module checksums
```

### Building Locally

1. Install dependencies:
```bash
go mod tidy
```

2. Build individual components:
```bash
go build -o bin/producer cmd/producer/main.go
go build -o bin/router cmd/router/main.go  
go build -o bin/consumer cmd/consumer/main.go
```

3. Run locally:
```bash
./bin/producer
./bin/router
./bin/consumer
```

## Monitoring

- **Kafka UI**: Access `http://localhost:8080` for a web-based interface
- **Docker Logs**: Use `docker-compose logs [service-name]` to view service logs
- **Consumer Output**: The consumer prints all received messages to stdout

## Use Cases

This ESB emulator is useful for:

- **Learning**: Understanding event-driven architecture patterns
- **Testing**: Simulating message flows in distributed systems
- **Development**: Prototyping Kafka-based applications
- **Demonstrations**: Showcasing ESB concepts and message routing

## Dependencies

- [segmentio/kafka-go](https://github.com/segmentio/kafka-go): Kafka client library
- [brianvoe/gofakeit](https://github.com/brianvoe/gofakeit): Fake data generation
- [Bitnami Kafka](https://hub.docker.com/r/bitnami/kafka): Kafka Docker image
- [Kafka UI](https://github.com/provectus/kafka-ui): Web UI for Kafka
