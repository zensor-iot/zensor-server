package internal

import (
	"fmt"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"database/sql/driver"
	"encoding/json"
)

type CommandSet []Command

func (CommandSet) TableName() string {
	return "device_commands"
}

func (s CommandSet) ToDomain() []domain.Command {
	result := make([]domain.Command, len(s))
	for i, v := range s {
		result[i] = v.ToDomain()
	}

	return result
}

type Command struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	Version       int        `json:"version"`
	DeviceName    string     `json:"device_name"`
	DeviceID      string     `json:"device_id"`
	TaskID        string     `json:"task_id" gorm:"foreignKey:task_id"`
	PayloadIndex  int        `json:"payload_index" gorm:"column:payload_index"`
	PayloadValue  int        `json:"payload_value" gorm:"column:payload_value"`
	DispatchAfter utils.Time `json:"dispatch_after"`
	Port          uint8      `json:"port"`
	Priority      string     `json:"priority"`
	CreatedAt     utils.Time `json:"created_at"`
	Ready         bool       `json:"ready"`
	Sent          bool       `json:"sent"`
	SentAt        utils.Time `json:"sent_at"`

	// Response tracking fields
	Status       string      `json:"status" gorm:"default:pending"`
	ErrorMessage *string     `json:"error_message,omitempty"`
	QueuedAt     *utils.Time `json:"queued_at,omitempty"`
	AckedAt      *utils.Time `json:"acked_at,omitempty"`
	FailedAt     *utils.Time `json:"failed_at,omitempty"`
}

func (Command) TableName() string {
	return "device_commands"
}

type CommandPayload struct {
	Index uint8 `json:"index"`
	Data  uint8 `json:"value"`
}

func (v CommandPayload) Value() (driver.Value, error) {
	return json.Marshal(v)
}

func (v *CommandPayload) Scan(value any) error {
	var data []byte

	switch val := value.(type) {
	case string:
		data = []byte(val)
	case []byte:
		data = val
	case nil:
		return nil
	default:
		return fmt.Errorf("cannot scan %T into CommandPayload", value)
	}

	return json.Unmarshal(data, &v)
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
			Index: domain.Index(c.PayloadIndex),
			Value: domain.CommandValue(c.PayloadValue),
		},
		DispatchAfter: c.DispatchAfter,
		CreatedAt:     c.CreatedAt,
		Ready:         c.Ready,
		Sent:          c.Sent,
		SentAt:        c.SentAt,

		// Response tracking fields
		Status:       domain.CommandStatus(c.Status),
		ErrorMessage: c.ErrorMessage,
		QueuedAt:     c.QueuedAt,
		AckedAt:      c.AckedAt,
		FailedAt:     c.FailedAt,
	}
}

func FromCommand(cmd domain.Command) Command {
	return Command{
		ID:            cmd.ID.String(),
		Version:       int(cmd.Version),
		DeviceID:      cmd.Device.ID.String(),
		DeviceName:    cmd.Device.Name,
		TaskID:        cmd.Task.ID.String(),
		PayloadIndex:  int(cmd.Payload.Index),
		PayloadValue:  int(cmd.Payload.Value),
		DispatchAfter: cmd.DispatchAfter,
		Port:          uint8(cmd.Port),
		Priority:      string(cmd.Priority),
		CreatedAt:     cmd.CreatedAt,
		Ready:         cmd.Ready,
		Sent:          cmd.Sent,
		SentAt:        cmd.SentAt,

		// Response tracking fields
		Status:       string(cmd.Status),
		ErrorMessage: cmd.ErrorMessage,
		QueuedAt:     cmd.QueuedAt,
		AckedAt:      cmd.AckedAt,
		FailedAt:     cmd.FailedAt,
	}
}
