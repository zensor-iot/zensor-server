package persistence

import (
	"context"
	"encoding/json"
	"time"
	"zensor-server/internal/infra/sql"
)

const (
	defaultRowLimit int = 10
)

type EventRepository interface {
	GetEvents() []*EventRecord
}

type EventRecord struct {
	ID       string    `json:"id"`
	DeviceID string    `json:"device_id"`
	Instant  time.Time `json:"instant"`
	Kind     string    `json:"kind"`
	Data     any       `json:"data"`
}

type eventRepositoryDefault struct {
	database sql.Database
}

func NewEventRepository(d sql.Database) EventRepository {
	return &eventRepositoryDefault{d}
}

func (r *eventRepositoryDefault) GetEvents() []*EventRecord {
	query := "SELECT * FROM events LIMIT $1"
	result, _ := r.database.Query(context.Background(), query, defaultRowLimit)
	events := unmarshalEventRecordArray(result)
	return events
}

func unmarshalEventRecordArray(data [][]byte) []*EventRecord {
	events := make([]*EventRecord, 0)
	for _, item := range data {
		event, _ := unmarshalEventRecord(item)
		events = append(events, event)
	}
	return events
}

func unmarshalEventRecord(data []byte) (*EventRecord, error) {
	var record EventRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}

	return &record, nil
}
