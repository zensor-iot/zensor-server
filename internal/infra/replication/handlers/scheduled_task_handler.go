package handlers

import (
	"context"
	"fmt"
	"reflect"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
)

// ScheduledTaskData represents the scheduled task table structure for GORM operations
type ScheduledTaskData struct {
	ID               string      `json:"id" gorm:"primaryKey"`
	Version          int         `json:"version"`
	TenantID         string      `json:"tenant_id" gorm:"index"`
	DeviceID         string      `json:"device_id" gorm:"index"`
	CommandTemplates string      `json:"command_templates"` // JSON array of command templates
	Schedule         string      `json:"schedule"`
	IsActive         bool        `json:"is_active"`
	CreatedAt        utils.Time  `json:"created_at"`
	UpdatedAt        utils.Time  `json:"updated_at"`
	LastExecutedAt   *utils.Time `json:"last_executed_at"`
	DeletedAt        *utils.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (ScheduledTaskData) TableName() string {
	return "scheduled_tasks_final"
}

// ScheduledTaskHandler handles replication of scheduled task data
type ScheduledTaskHandler struct {
	orm sql.ORM
}

// NewScheduledTaskHandler creates a new scheduled task handler
func NewScheduledTaskHandler(orm sql.ORM) *ScheduledTaskHandler {
	return &ScheduledTaskHandler{
		orm: orm,
	}
}

// TopicName returns the scheduled_tasks topic
func (h *ScheduledTaskHandler) TopicName() pubsub.Topic {
	return "scheduled_tasks"
}

// Create handles creating a new scheduled task record
func (h *ScheduledTaskHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	internalScheduledTask := h.extractScheduledTaskFields(message)

	err := h.orm.WithContext(ctx).Create(&internalScheduledTask).Error()
	if err != nil {
		return fmt.Errorf("creating scheduled task: %w", err)
	}

	return nil
}

// GetByID retrieves a scheduled task by its ID
func (h *ScheduledTaskHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var internalScheduledTask ScheduledTaskData

	err := h.orm.WithContext(ctx).First(&internalScheduledTask, "id = ?", id).Error()
	if err != nil {
		return nil, fmt.Errorf("getting scheduled task: %w", err)
	}

	scheduledTask := h.toDomainScheduledTask(internalScheduledTask)
	return scheduledTask, nil
}

// Update handles updating an existing scheduled task record
func (h *ScheduledTaskHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	var existing ScheduledTaskData
	if err := h.orm.WithContext(ctx).First(&existing, "id = ?", string(key)).Error(); err != nil {
		return fmt.Errorf("fetching existing scheduled task: %w", err)
	}

	incoming := h.extractScheduledTaskFields(message)

	// Update fields if they are not empty
	if incoming.TenantID != "" {
		existing.TenantID = incoming.TenantID
	}
	if incoming.DeviceID != "" {
		existing.DeviceID = incoming.DeviceID
	}
	if incoming.CommandTemplates != "" {
		existing.CommandTemplates = incoming.CommandTemplates
	}
	if incoming.Schedule != "" {
		existing.Schedule = incoming.Schedule
	}
	existing.IsActive = incoming.IsActive
	existing.UpdatedAt = incoming.UpdatedAt
	existing.Version = incoming.Version
	existing.LastExecutedAt = incoming.LastExecutedAt
	existing.DeletedAt = incoming.DeletedAt

	if err := h.orm.WithContext(ctx).Save(&existing).Error(); err != nil {
		return fmt.Errorf("updating scheduled task: %w", err)
	}
	return nil
}

// extractScheduledTaskFields uses reflection to extract scheduled task fields from any message type
func (h *ScheduledTaskHandler) extractScheduledTaskFields(message pubsub.Message) ScheduledTaskData {
	val := reflect.ValueOf(message)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := ScheduledTaskData{
		Version: 1,
	}

	if idField := val.FieldByName("ID"); idField.IsValid() {
		result.ID = idField.Interface().(string)
	}

	if versionField := val.FieldByName("Version"); versionField.IsValid() {
		versionInterface := versionField.Interface()
		switch v := versionInterface.(type) {
		case int:
			result.Version = v
		case uint:
			result.Version = int(v)
		default:
			result.Version = 1
		}
	}

	if tenantIDField := val.FieldByName("TenantID"); tenantIDField.IsValid() {
		result.TenantID = tenantIDField.Interface().(string)
	}

	if deviceIDField := val.FieldByName("DeviceID"); deviceIDField.IsValid() {
		result.DeviceID = deviceIDField.Interface().(string)
	}

	if commandTemplatesField := val.FieldByName("CommandTemplates"); commandTemplatesField.IsValid() {
		result.CommandTemplates = commandTemplatesField.Interface().(string)
	}

	if scheduleField := val.FieldByName("Schedule"); scheduleField.IsValid() {
		result.Schedule = scheduleField.Interface().(string)
	}

	if isActiveField := val.FieldByName("IsActive"); isActiveField.IsValid() {
		result.IsActive = isActiveField.Interface().(bool)
	}

	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		createdAtInterface := createdAtField.Interface()
		switch v := createdAtInterface.(type) {
		case utils.Time:
			result.CreatedAt = v
		case time.Time:
			result.CreatedAt = utils.Time{Time: v}
		default:
			// Default to current time if we can't convert
			result.CreatedAt = utils.Time{Time: time.Now()}
		}
	}

	if updatedAtField := val.FieldByName("UpdatedAt"); updatedAtField.IsValid() {
		updatedAtInterface := updatedAtField.Interface()
		switch v := updatedAtInterface.(type) {
		case utils.Time:
			result.UpdatedAt = v
		case time.Time:
			result.UpdatedAt = utils.Time{Time: v}
		default:
			// Default to current time if we can't convert
			result.UpdatedAt = utils.Time{Time: time.Now()}
		}
	}

	if lastExecutedAtField := val.FieldByName("LastExecutedAt"); lastExecutedAtField.IsValid() {
		if lastExecutedAtField.IsNil() {
			result.LastExecutedAt = nil
		} else {
			lastExecutedAt := lastExecutedAtField.Interface().(*utils.Time)
			result.LastExecutedAt = lastExecutedAt
		}
	}

	if deletedAtField := val.FieldByName("DeletedAt"); deletedAtField.IsValid() {
		if deletedAtField.IsNil() {
			result.DeletedAt = nil
		} else {
			deletedAt := deletedAtField.Interface().(*utils.Time)
			result.DeletedAt = deletedAt
		}
	}

	return result
}

func (h *ScheduledTaskHandler) toDomainScheduledTask(internalScheduledTask ScheduledTaskData) map[string]any {
	return map[string]any{
		"id":                internalScheduledTask.ID,
		"version":           internalScheduledTask.Version,
		"tenant_id":         internalScheduledTask.TenantID,
		"device_id":         internalScheduledTask.DeviceID,
		"command_templates": internalScheduledTask.CommandTemplates,
		"schedule":          internalScheduledTask.Schedule,
		"is_active":         internalScheduledTask.IsActive,
		"created_at":        internalScheduledTask.CreatedAt,
		"updated_at":        internalScheduledTask.UpdatedAt,
		"last_executed_at":  internalScheduledTask.LastExecutedAt,
		"deleted_at":        internalScheduledTask.DeletedAt,
	}
}
