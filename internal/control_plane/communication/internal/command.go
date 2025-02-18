package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
)

type Command struct {
	DeviceName string         `json:"device_name"`
	DeviceID   string         `json:"device_id"`
	Payload    CommandPayload `json:"payload"`
	Port       uint8          `json:"port"`
	Priority   string         `json:"priority"`
	CreatedAt  time.Time      `json:"created_at"`
}

type CommandPayload struct {
	Index uint8 `json:"index"`
	Value uint8 `json:"value"`
}

func FromCommand(cmd domain.Command) Command {
	return Command{
		DeviceID:   cmd.Device.ID.String(),
		DeviceName: cmd.Device.Name,
		Payload: CommandPayload{
			Index: uint8(cmd.Payload.Index),
			Value: uint8(cmd.Payload.Value),
		},
		Port:      uint8(cmd.Port),
		Priority:  string(cmd.Priority),
		CreatedAt: time.Now(),
	}
}
