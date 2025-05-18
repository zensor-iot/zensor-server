package utils

import (
	"encoding/json"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*d = Duration(val)
	return nil
}

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	formatted := t.UTC().Format("2006-01-02T15:04:05.000Z07:00")
	return []byte(`"` + formatted + `"`), nil
}
