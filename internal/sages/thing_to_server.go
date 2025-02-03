package sages

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"zensor-server/internal/infra/kafka"
	"zensor-server/internal/infra/mqtt"
)

const (
	eventKindGeneric = "generic"
)

var (
	myID string
)

type DeviceRegisteredEvent struct {
	DeviceID     string    `json:"device_id"`
	RegisteredAt time.Time `json:"registered_at"`
}

type DeviceEvent struct {
	ID       string      `json:"id"`
	DeviceID string      `json:"device_id"`
	Instant  time.Time   `json:"instant"`
	Kind     string      `json:"kind"`
	Data     interface{} `json:"data"`
}

func EvaluateThingToServer(events chan mqtt.Event, publisher kafka.KafkaPublisher, eventPublisher kafka.KafkaPublisher) {
	for event := range events {
		slog.Info("message received", slog.Any("event", event))
		switch event.Type {
		case mqtt.EventTypeMyself:
			myID = event.DeviceID
		case mqtt.EventTypePresence:
			evaluatePresence(event, publisher)
		case mqtt.EventTypeMessage:
			evaluateEvent(event, eventPublisher)
		default:
			slog.Info("event type not supported", slog.String("type", event.Type))
		}
	}
}

func evaluatePresence(e mqtt.Event, p kafka.KafkaPublisher) {
	if e.DeviceID == myID {
		slog.Info("event discarded given comes from myself")
		return
	}

	val, _ := json.Marshal(&DeviceRegisteredEvent{
		e.DeviceID,
		time.Now(),
	})
	p.Publish(context.Background(), e.DeviceID, string(val))
}

func evaluateEvent(e mqtt.Event, p kafka.KafkaPublisher) {
	val, _ := json.Marshal(&DeviceEvent{
		ID:       uuid.NewString(),
		DeviceID: e.DeviceID,
		Instant:  time.Now(),
		Kind:     eventKindGeneric,
		Data:     string(e.Value),
	})
	p.Publish(context.Background(), e.DeviceID, string(val))
}
