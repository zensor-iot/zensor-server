package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
)

type Command struct {
	DeviceName string    `json:"device_name"`
	DeviceID   string    `json:"device_id"`
	RawPayload string    `json:"raw_payload"`
	Port       uint8     `json:"port"`
	Priority   string    `json:"priority"`
	CreatedAt  time.Time `json:"created_at"`
}

func FromCommand(cmd domain.Command) Command {
	return Command{
		DeviceID:   cmd.Device.ID.String(),
		DeviceName: cmd.Device.Name,
		RawPayload: cmd.RawPayload,
		Port:       cmd.Port,
		Priority:   cmd.Priority,
		CreatedAt:  time.Now(),
	}
}
