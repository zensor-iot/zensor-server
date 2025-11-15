package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/avro"
)

type MaintenanceActivityData struct {
	ID                     string      `json:"id" gorm:"primaryKey"`
	Version                int         `json:"version"`
	TenantID               string      `json:"tenant_id" gorm:"index;not null"`
	TypeName               string      `json:"type_name" gorm:"not null"`
	CustomTypeName         *string     `json:"custom_type_name,omitempty"`
	Name                   string      `json:"name" gorm:"not null"`
	Description            string      `json:"description"`
	Schedule               string      `json:"schedule" gorm:"not null"`
	NotificationDaysBefore string      `json:"notification_days_before" gorm:"type:text"`
	Fields                 string      `json:"fields" gorm:"type:text"`
	IsActive               bool        `json:"is_active" gorm:"default:true"`
	CreatedAt              utils.Time  `json:"created_at"`
	UpdatedAt              utils.Time  `json:"updated_at"`
	DeletedAt              *utils.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (MaintenanceActivityData) TableName() string {
	return "maintenance_activities"
}

type MaintenanceActivityHandler struct {
	orm sql.ORM
}

func NewMaintenanceActivityHandler(orm sql.ORM) *MaintenanceActivityHandler {
	return &MaintenanceActivityHandler{
		orm: orm,
	}
}

func (h *MaintenanceActivityHandler) TopicName() pubsub.Topic {
	return "maintenance_activities"
}

func (h *MaintenanceActivityHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	internalActivity := h.extractMaintenanceActivityFields(message)

	err := h.orm.WithContext(ctx).Create(&internalActivity).Error()
	if err != nil {
		return fmt.Errorf("creating maintenance activity: %w", err)
	}

	return nil
}

func (h *MaintenanceActivityHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var internalActivity MaintenanceActivityData

	err := h.orm.WithContext(ctx).First(&internalActivity, "id = ?", id).Error()
	if err != nil {
		return nil, fmt.Errorf("getting maintenance activity: %w", err)
	}

	activity := h.toDomainMaintenanceActivity(internalActivity)
	return activity, nil
}

func (h *MaintenanceActivityHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	var existing MaintenanceActivityData
	if err := h.orm.WithContext(ctx).First(&existing, "id = ?", string(key)).Error(); err != nil {
		return fmt.Errorf("fetching existing maintenance activity: %w", err)
	}
	incoming := h.extractMaintenanceActivityFields(message)
	existing.Name = incoming.Name
	existing.Description = incoming.Description
	existing.Schedule = incoming.Schedule
	existing.NotificationDaysBefore = incoming.NotificationDaysBefore
	existing.Fields = incoming.Fields
	existing.IsActive = incoming.IsActive
	existing.UpdatedAt = incoming.UpdatedAt
	existing.DeletedAt = incoming.DeletedAt
	existing.Version = incoming.Version

	if err := h.orm.WithContext(ctx).Save(&existing).Error(); err != nil {
		return fmt.Errorf("updating maintenance activity: %w", err)
	}
	return nil
}

func (h *MaintenanceActivityHandler) extractMaintenanceActivityFields(message pubsub.Message) MaintenanceActivityData {
	avroActivity, ok := message.(*avro.AvroMaintenanceActivity)
	if !ok {
		slog.Error("message is not *avro.AvroMaintenanceActivity", "message", message)
		return MaintenanceActivityData{}
	}

	notificationDaysBeforeBytes, _ := json.Marshal(avroActivity.NotificationDaysBefore)
	notificationDaysBefore := string(notificationDaysBeforeBytes)

	fieldsBytes, _ := json.Marshal(avroActivity.Fields)
	fields := string(fieldsBytes)

	var deletedAt *utils.Time
	if avroActivity.DeletedAt != nil {
		deletedAt = &utils.Time{Time: *avroActivity.DeletedAt}
	}

	return MaintenanceActivityData{
		ID:                     avroActivity.ID,
		Version:                avroActivity.Version,
		TenantID:               avroActivity.TenantID,
		TypeName:               avroActivity.TypeName,
		CustomTypeName:         avroActivity.CustomTypeName,
		Name:                   avroActivity.Name,
		Description:            avroActivity.Description,
		Schedule:               avroActivity.Schedule,
		NotificationDaysBefore: notificationDaysBefore,
		Fields:                 fields,
		IsActive:               avroActivity.IsActive,
		CreatedAt:              utils.Time{Time: avroActivity.CreatedAt},
		UpdatedAt:              utils.Time{Time: avroActivity.UpdatedAt},
		DeletedAt:              deletedAt,
	}
}

func (h *MaintenanceActivityHandler) toDomainMaintenanceActivity(internalActivity MaintenanceActivityData) map[string]any {
	return map[string]any{
		"id":   internalActivity.ID,
		"name": internalActivity.Name,
	}
}
