package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"
)

type Command struct {
	ID            string         `json:"id"`
	Version       int            `json:"version"`
	DeviceName    string         `json:"device_name"`
	DeviceID      string         `json:"device_id"`
	TaskID        string         `json:"task_id"`
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
	Index uint8 `json:"index"`
	Value uint8 `json:"value"`
}

func FromCommand(cmd domain.Command) Command {
	return Command{
		ID:         cmd.ID.String(),
		Version:    int(cmd.Version),
		DeviceID:   cmd.Device.ID.String(),
		TaskID:     cmd.Task.ID.String(),
		DeviceName: cmd.Device.Name,
		Payload: CommandPayload{
			Index: uint8(cmd.Payload.Index),
			Value: uint8(cmd.Payload.Value),
		},
		DispatchAfter: cmd.DispatchAfter,
		Port:          uint8(cmd.Port),
		Priority:      string(cmd.Priority),
		CreatedAt:     utils.Time{Time: time.Now()},
		Ready:         cmd.Ready,
		Sent:          cmd.Sent,
		SentAt:        cmd.SentAt,
	}
}
