package handlers

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/avro"
)

type CommandData struct {
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

	// Response tracking fields
	Status       string      `json:"status" gorm:"default:pending"`
	ErrorMessage *string     `json:"error_message,omitempty"`
	QueuedAt     *utils.Time `json:"queued_at,omitempty"`
	AckedAt      *utils.Time `json:"acked_at,omitempty"`
	FailedAt     *utils.Time `json:"failed_at,omitempty"`
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

func (CommandData) TableName() string {
	return "device_commands"
}

type CommandHandler struct {
	orm sql.ORM
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(orm sql.ORM) *CommandHandler {
	return &CommandHandler{
		orm: orm,
	}
}

// TopicName returns the commands topic
func (h *CommandHandler) TopicName() pubsub.Topic {
	return "device_commands"
}

// Create handles creating a new command record
func (h *CommandHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	internalCommand := h.extractCommandFields(message)

	err := h.orm.WithContext(ctx).Create(&internalCommand).Error()
	if err != nil {
		return fmt.Errorf("creating command: %w", err)
	}

	return nil
}

// GetByID retrieves a command by its ID
func (h *CommandHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var internalCommand CommandData

	err := h.orm.WithContext(ctx).First(&internalCommand, "id = ?", id).Error()
	if err != nil {
		return nil, fmt.Errorf("getting command: %w", err)
	}

	command := h.toDomainCommand(internalCommand)
	return command, nil
}

// Update handles updating an existing command record
func (h *CommandHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	internalCommand := h.extractCommandFields(message)

	// First, fetch the existing record to preserve created_at
	var existingCommand CommandData
	err := h.orm.WithContext(ctx).First(&existingCommand, "id = ?", internalCommand.ID).Error()
	if err != nil {
		return fmt.Errorf("fetching existing command: %w", err)
	}

	// Update only the fields that should change, preserving created_at
	existingCommand.Version = internalCommand.Version
	existingCommand.DeviceName = internalCommand.DeviceName
	existingCommand.DeviceID = internalCommand.DeviceID
	existingCommand.TaskID = internalCommand.TaskID
	existingCommand.Payload = internalCommand.Payload
	existingCommand.DispatchAfter = internalCommand.DispatchAfter
	existingCommand.Port = internalCommand.Port
	existingCommand.Priority = internalCommand.Priority
	existingCommand.Ready = internalCommand.Ready
	existingCommand.Sent = internalCommand.Sent
	existingCommand.SentAt = internalCommand.SentAt
	existingCommand.Status = internalCommand.Status
	existingCommand.ErrorMessage = internalCommand.ErrorMessage
	existingCommand.QueuedAt = internalCommand.QueuedAt
	existingCommand.AckedAt = internalCommand.AckedAt
	existingCommand.FailedAt = internalCommand.FailedAt

	// Save the updated record
	err = h.orm.WithContext(ctx).Save(&existingCommand).Error()
	if err != nil {
		return fmt.Errorf("updating command: %w", err)
	}

	return nil
}

// extractCommandFields uses reflection to extract command fields from any message type
func (h *CommandHandler) extractCommandFields(message pubsub.Message) CommandData {
	avroCommand, ok := message.(*avro.AvroCommand)
	if !ok {
		slog.Error("message is not *avro.AvroCommand", "message", message)
		return CommandData{}
	}

	return CommandData{
		ID:         avroCommand.ID,
		Version:    int(avroCommand.Version),
		DeviceName: avroCommand.DeviceName,
		DeviceID:   avroCommand.DeviceID,
		TaskID:     avroCommand.TaskID,
		Payload: CommandPayload{
			Index: uint8(avroCommand.PayloadIndex),
			Data:  uint8(avroCommand.PayloadValue),
		},
		DispatchAfter: utils.Time{Time: avroCommand.DispatchAfter},
		Port:          uint8(avroCommand.Port),
		Priority:      avroCommand.Priority,
		CreatedAt:     utils.Time{Time: avroCommand.CreatedAt},
		Ready:         avroCommand.Ready,
		Sent:          avroCommand.Sent,
		SentAt:        utils.Time{Time: avroCommand.SentAt},

		// Response tracking fields
		Status:       avroCommand.Status,
		ErrorMessage: avroCommand.ErrorMessage,
		QueuedAt:     convertTimePtr(avroCommand.QueuedAt),
		AckedAt:      convertTimePtr(avroCommand.AckedAt),
		FailedAt:     convertTimePtr(avroCommand.FailedAt),
	}
}

// convertTimePtr converts a *time.Time to *utils.Time
func convertTimePtr(t *time.Time) *utils.Time {
	if t == nil {
		return nil
	}
	return &utils.Time{Time: *t}
}

func (h *CommandHandler) toDomainCommand(internalCommand CommandData) map[string]any {
	return map[string]any{
		"id":          internalCommand.ID,
		"device_name": internalCommand.DeviceName,
		"device_id":   internalCommand.DeviceID,
		"task_id":     internalCommand.TaskID,
		"payload": map[string]any{
			"index": internalCommand.Payload.Index,
			"data":  internalCommand.Payload.Data,
		},
		"dispatch_after": internalCommand.DispatchAfter,
		"port":           internalCommand.Port,
		"priority":       internalCommand.Priority,
		"created_at":     internalCommand.CreatedAt,
		"ready":          internalCommand.Ready,
		"sent":           internalCommand.Sent,
		"sent_at":        internalCommand.SentAt,
	}
}
