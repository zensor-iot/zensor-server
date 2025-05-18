package shared_kernel

import (
	"fmt"

	"zensor-server/internal/infra/utils"

	"github.com/vmihailenco/msgpack/v5"
)

type Command struct {
	ID            string         `json:"id"`
	DeviceID      string         `json:"device_id"`
	DeviceName    string         `json:"device_name"`
	Payload       CommandPayload `json:"payload"`
	DispatchAfter utils.Time     `json:"dispatch_after"`
	Port          uint8          `json:"port"`
	Priority      string         `json:"priority"`
	CreatedAt     utils.Time     `json:"created_at"`
	Ready         bool           `json:"ready"`
	Sent          bool           `json:"sent"`
	SentAt        utils.Time     `json:"sent_at"`
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
