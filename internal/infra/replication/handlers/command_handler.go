package handlers

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
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
	return "device_commands_final"
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
	fmt.Printf("*** Creating command: %+v\n", message)
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

	err := h.orm.WithContext(ctx).Save(&internalCommand).Error()
	if err != nil {
		return fmt.Errorf("updating command: %w", err)
	}

	return nil
}

// extractCommandFields uses reflection to extract command fields from any message type
func (h *CommandHandler) extractCommandFields(message pubsub.Message) CommandData {
	val := reflect.ValueOf(message)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := CommandData{
		Version: 1,
	}

	if idField := val.FieldByName("ID"); idField.IsValid() {
		result.ID = idField.Interface().(string)
	}

	if deviceNameField := val.FieldByName("DeviceName"); deviceNameField.IsValid() {
		result.DeviceName = deviceNameField.Interface().(string)
	}

	if deviceIDField := val.FieldByName("DeviceID"); deviceIDField.IsValid() {
		result.DeviceID = deviceIDField.Interface().(string)
	}

	if taskIDField := val.FieldByName("TaskID"); taskIDField.IsValid() {
		result.TaskID = taskIDField.Interface().(string)
	}

	if payloadField := val.FieldByName("Payload"); payloadField.IsValid() {
		payload := payloadField.Interface()
		payloadVal := reflect.ValueOf(payload)
		if payloadVal.Kind() == reflect.Ptr {
			payloadVal = payloadVal.Elem()
		}

		if indexField := payloadVal.FieldByName("Index"); indexField.IsValid() {
			result.Payload.Index = uint8(indexField.Interface().(uint8))
		}

		if dataField := payloadVal.FieldByName("Data"); dataField.IsValid() {
			result.Payload.Data = uint8(dataField.Interface().(uint8))
		}
	}

	if dispatchAfterField := val.FieldByName("DispatchAfter"); dispatchAfterField.IsValid() {
		result.DispatchAfter = dispatchAfterField.Interface().(utils.Time)
	}

	if portField := val.FieldByName("Port"); portField.IsValid() {
		result.Port = uint8(portField.Interface().(uint8))
	}

	if priorityField := val.FieldByName("Priority"); priorityField.IsValid() {
		result.Priority = priorityField.Interface().(string)
	}

	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		result.CreatedAt = createdAtField.Interface().(utils.Time)
	}

	if readyField := val.FieldByName("Ready"); readyField.IsValid() {
		result.Ready = readyField.Interface().(bool)
	}

	if sentField := val.FieldByName("Sent"); sentField.IsValid() {
		result.Sent = sentField.Interface().(bool)
	}

	if sentAtField := val.FieldByName("SentAt"); sentAtField.IsValid() {
		result.SentAt = sentAtField.Interface().(utils.Time)
	}

	return result
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
