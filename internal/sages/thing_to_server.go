package sages

import (
	"encoding/json"
	"time"

	"zensor-server/internal/logger"

	"github.com/google/uuid"

	"zensor-server/internal/kafka"
	"zensor-server/internal/mqtt"
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
		logger.Info("message received", "event", event)
		switch event.Type {
		case mqtt.EventTypeMyself:
			myID = event.DeviceID
		case mqtt.EventTypePresence:
			evaluatePresence(event, publisher)
		case mqtt.EventTypeMessage:
			evaluateEvent(event, eventPublisher)
		default:
			logger.Info("event type %s not supported", event.Type)
		}
	}
}

func evaluatePresence(e mqtt.Event, p kafka.KafkaPublisher) {
	if e.DeviceID == myID {
		logger.Info("event discarded given comes from myself")
		return
	}

	val, _ := json.Marshal(&DeviceRegisteredEvent{
		e.DeviceID,
		time.Now(),
	})
	p.Publish(e.DeviceID, string(val))
}

func evaluateEvent(e mqtt.Event, p kafka.KafkaPublisher) {
	val, _ := json.Marshal(&DeviceEvent{
		ID:       uuid.NewString(),
		DeviceID: e.DeviceID,
		Instant:  time.Now(),
		Kind:     eventKindGeneric,
		Data:     string(e.Value),
	})
	p.Publish(e.DeviceID, string(val))
}
