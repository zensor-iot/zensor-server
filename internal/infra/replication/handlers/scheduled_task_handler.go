package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/avro"
)

// ScheduledTaskData represents the scheduled task table structure for GORM operations
type ScheduledTaskData struct {
	ID               string      `json:"id" gorm:"primaryKey"`
	Version          int         `json:"version"`
	TenantID         string      `json:"tenant_id" gorm:"index"`
	DeviceID         string      `json:"device_id" gorm:"index"`
	CommandTemplates string      `json:"command_templates"` // JSON array of command templates
	Schedule         string      `json:"schedule"`
	SchedulingConfig string      `json:"scheduling_config"` // JSON scheduling configuration
	IsActive         bool        `json:"is_active"`
	CreatedAt        utils.Time  `json:"created_at"`
	UpdatedAt        utils.Time  `json:"updated_at"`
	LastExecutedAt   *utils.Time `json:"last_executed_at"`
	DeletedAt        *utils.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (ScheduledTaskData) TableName() string {
	return "scheduled_tasks"
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
	if incoming.SchedulingConfig != "" {
		existing.SchedulingConfig = incoming.SchedulingConfig
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
	avroScheduledTask, ok := message.(*avro.AvroScheduledTask)
	if !ok {
		slog.Error("message is not *avro.AvroScheduledTask", "message", message)
		return ScheduledTaskData{}
	}

	result := ScheduledTaskData{
		ID:               avroScheduledTask.ID,
		Version:          int(avroScheduledTask.Version),
		TenantID:         avroScheduledTask.TenantID,
		DeviceID:         avroScheduledTask.DeviceID,
		CommandTemplates: avroScheduledTask.CommandTemplates,
		Schedule:         avroScheduledTask.Schedule,
		IsActive:         avroScheduledTask.IsActive,
		CreatedAt:        utils.Time{Time: avroScheduledTask.CreatedAt},
		UpdatedAt:        utils.Time{Time: avroScheduledTask.UpdatedAt},
	}

	if avroScheduledTask.SchedulingConfig != nil {
		result.SchedulingConfig = *avroScheduledTask.SchedulingConfig
	}

	if avroScheduledTask.LastExecutedAt != nil {
		result.LastExecutedAt = &utils.Time{Time: *avroScheduledTask.LastExecutedAt}
	}
	if avroScheduledTask.DeletedAt != nil {
		result.DeletedAt = &utils.Time{Time: *avroScheduledTask.DeletedAt}
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
		"scheduling_config": internalScheduledTask.SchedulingConfig,
		"is_active":         internalScheduledTask.IsActive,
		"created_at":        internalScheduledTask.CreatedAt,
		"updated_at":        internalScheduledTask.UpdatedAt,
		"last_executed_at":  internalScheduledTask.LastExecutedAt,
		"deleted_at":        internalScheduledTask.DeletedAt,
	}
}
