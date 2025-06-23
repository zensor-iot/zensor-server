package internal

import (
	"errors"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"

	"database/sql/driver"
	"encoding/json"
)

type CommandSet []Command

func (CommandSet) TableName() string {
	return "device_commands_final"
}

func (s CommandSet) ToDomain() []domain.Command {
	result := make([]domain.Command, len(s))
	for i, v := range s {
		result[i] = v.ToDomain()
	}

	return result
}

type Command struct {
	ID            string         `json:"id" gorm:"primaryKey"`
	Version       int            `json:"version"`
	DeviceName    string         `json:"device_name"`
	DeviceID      string         `json:"device_id"`
	TaskID        string         `json:"task_id" gorm:"foreignKey:task_id"`
	Payload       CommandPayload `json:"payload" gorm:"type:json"`
	DispatchAfter utils.Time     `json:"dispatch_after"`
	Port          uint8          `json:"port"`
	Priority      string         `json:"priority"`
	CreatedAt     utils.Time     `json:"created_at"`
	Ready         bool           `json:"ready"`
	Sent          bool           `json:"sent"`
	SentAt        utils.Time     `json:"sent_at"`
}

func (Command) TableName() string {
	return "device_commands_final"
}

type CommandPayload struct {
	Index uint8 `json:"index"`
	Data  uint8 `json:"value"`
}

func (v CommandPayload) Value() (driver.Value, error) {
	return json.Marshal(v)
}

func (v *CommandPayload) Scan(value any) error {
	data, ok := value.(string)
	if !ok {
		return errors.New("type assertion to string failed")
	}
	return json.Unmarshal([]byte(data), &v)
}

func (c Command) ToDomain() domain.Command {
	return domain.Command{
		ID:       domain.ID(c.ID),
		Version:  domain.Version(c.Version),
		Device:   domain.Device{ID: domain.ID(c.DeviceID), Name: c.DeviceName},
		Task:     domain.Task{ID: domain.ID(c.TaskID)},
		Port:     domain.Port(c.Port),
		Priority: domain.CommandPriority(c.Priority),
		Payload: domain.CommandPayload{
			Index: domain.Index(c.Payload.Index),
			Value: domain.CommandValue(c.Payload.Data),
		},
		DispatchAfter: c.DispatchAfter,
		Ready:         c.Ready,
		Sent:          c.Sent,
		SentAt:        c.SentAt,
	}
}

func FromCommand(cmd domain.Command) Command {
	return Command{
		ID:         cmd.ID.String(),
		Version:    int(cmd.Version),
		DeviceID:   cmd.Device.ID.String(),
		DeviceName: cmd.Device.Name,
		Payload: CommandPayload{
			Index: uint8(cmd.Payload.Index),
			Data:  uint8(cmd.Payload.Value),
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
