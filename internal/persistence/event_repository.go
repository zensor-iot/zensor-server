package persistence

import (
	"encoding/json"
	"time"
)

const (
	defaultRowLimit int = 10
)

type EventRepository interface {
	GetEvents() []*EventRecord
}

type EventRecord struct {
	ID       string      `json:"id"`
	DeviceID string      `json:"device_id"`
	Instant  time.Time   `json:"instant"`
	Kind     string      `json:"kind"`
	Data     interface{} `json:"data"`
}

type eventRepositoryDefault struct {
	database Database
}

func NewEventRepository(d Database) EventRepository {
	return &eventRepositoryDefault{d}
}

func (r *eventRepositoryDefault) GetEvents() []*EventRecord {
	query := "SELECT * FROM events LIMIT $1"
	result := r.database.Query(query, defaultRowLimit)
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
