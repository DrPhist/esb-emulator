// generator.go

package generator

import (
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

type Event struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // "order" | "payment" | other
	TS       time.Time              `json:"ts"`
	Metadata map[string]string      `json:"metadata"`
	Payload  map[string]interface{} `json:"payload"`
}

func init() {
	seed := time.Now().UnixNano()
	gofakeit.Seed(seed)
}

func randomType() string {
	// Mostly valid, sometimes junk to exercise DLQ
	switch n := rand.Intn(10); {
	case n < 5:
		return "order"
	case n < 9:
		return "payment"
	default:
		return "mystery" // will route to DLQ
	}
}

func NewEvent() Event {
	t := randomType()
	id := gofakeit.UUID()
	meta := map[string]string{
		"correlation_id": gofakeit.UUID(),
		"source":         "emulator.v1",
	}
	payload := map[string]interface{}{}

	switch t {
	case "order":
		payload = map[string]interface{}{
			"order_id": id,
			"customer": gofakeit.Name(),
			"amount":   gofakeit.Price(10, 2000),
			"items":    gofakeit.Number(1, 5),
		}
	case "payment":
		payload = map[string]interface{}{
			"payment_id": id,
			"method":     gofakeit.RandomString([]string{"card", "ach", "wire"}),
			"amount":     gofakeit.Price(10, 2000),
			"currency":   gofakeit.CurrencyShort(),
		}
	}

	return Event{
		ID:       id,
		Type:     t,
		TS:       time.Now().UTC(),
		Metadata: meta,
		Payload:  payload,
	}
}
