package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"

	"encoding/json"
	"database/sql/driver"
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
	DeviceName    string         `json:"device_name"`
	DeviceID      string         `json:"device_id"`
	Payload       CommandPayload `json:"payload" gorm:"type:json"`
	DispatchAfter time.Time      `json:"dispatch_after"`
	Port          uint8          `json:"port"`
	Priority      string         `json:"priority"`
	CreatedAt     time.Time      `json:"created_at"`
	Ready         bool           `json:"ready"`
	Sent          bool           `json:"sent"`
	SentAt        time.Time      `json:"sent_at"`
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

func (c Command) ToDomain() domain.Command {
	return domain.Command{
		ID:       domain.ID(c.ID),
		Device:   domain.Device{ID: domain.ID(c.DeviceID), Name: c.DeviceName},
		Port:     domain.Port(c.Port),
		Priority: domain.CommandPriority(c.Priority),
		Payload: domain.CommandPayload{
			Index: domain.Index(c.Payload.Index),
			Value: domain.CommandValue(c.Payload.Data),
		},
		DispatchAfter: utils.Time{Time: c.DispatchAfter},
		Ready:         c.Ready,
		Sent:          c.Sent,
		SentAt:        utils.Time{Time: c.SentAt},
	}
}
