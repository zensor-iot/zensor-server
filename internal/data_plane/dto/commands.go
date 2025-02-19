package dto

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

type Command struct {
	DeviceID   string         `json:"device_id"`
	DeviceName string         `json:"device_name"`
	Payload    CommandPayload `json:"payload"`
	Port       uint8          `json:"port"`
	Priority   string         `json:"priority"`
}

type CommandPayload struct {
	Index uint8 `json:"index" msgpack:"i"`
	Value uint8 `json:"value" msgpack:"v"`
}

func (p CommandPayload) ToMessagePack() ([]byte, error) {
	data, err := msgpack.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("msgpack marshaling: %w", err)
	}
	return data, nil
}
