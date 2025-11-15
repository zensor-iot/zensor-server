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

type MaintenanceExecutionData struct {
	ID            string      `json:"id" gorm:"primaryKey"`
	Version       int         `json:"version"`
	ActivityID    string      `json:"activity_id" gorm:"index;not null"`
	ScheduledDate utils.Time  `json:"scheduled_date" gorm:"not null"`
	CompletedAt   *utils.Time `json:"completed_at,omitempty"`
	CompletedBy   *string     `json:"completed_by,omitempty"`
	OverdueDays   int         `json:"overdue_days" gorm:"default:0"`
	FieldValues   string      `json:"field_values" gorm:"type:text"`
	CreatedAt     utils.Time  `json:"created_at"`
	UpdatedAt     utils.Time  `json:"updated_at"`
	DeletedAt     *utils.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (MaintenanceExecutionData) TableName() string {
	return "maintenance_executions"
}

type MaintenanceExecutionHandler struct {
	orm sql.ORM
}

func NewMaintenanceExecutionHandler(orm sql.ORM) *MaintenanceExecutionHandler {
	return &MaintenanceExecutionHandler{
		orm: orm,
	}
}

func (h *MaintenanceExecutionHandler) TopicName() pubsub.Topic {
	return "maintenance_executions"
}

func (h *MaintenanceExecutionHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	internalExecution := h.extractMaintenanceExecutionFields(message)

	err := h.orm.WithContext(ctx).Create(&internalExecution).Error()
	if err != nil {
		return fmt.Errorf("creating maintenance execution: %w", err)
	}

	return nil
}

func (h *MaintenanceExecutionHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var internalExecution MaintenanceExecutionData

	err := h.orm.WithContext(ctx).First(&internalExecution, "id = ?", id).Error()
	if err != nil {
		return nil, fmt.Errorf("getting maintenance execution: %w", err)
	}

	execution := h.toDomainMaintenanceExecution(internalExecution)
	return execution, nil
}

func (h *MaintenanceExecutionHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	var existing MaintenanceExecutionData
	if err := h.orm.WithContext(ctx).First(&existing, "id = ?", string(key)).Error(); err != nil {
		return fmt.Errorf("fetching existing maintenance execution: %w", err)
	}
	incoming := h.extractMaintenanceExecutionFields(message)
	existing.ActivityID = incoming.ActivityID
	existing.ScheduledDate = incoming.ScheduledDate
	existing.CompletedAt = incoming.CompletedAt
	existing.CompletedBy = incoming.CompletedBy
	existing.OverdueDays = incoming.OverdueDays
	existing.FieldValues = incoming.FieldValues
	existing.UpdatedAt = incoming.UpdatedAt
	existing.DeletedAt = incoming.DeletedAt
	existing.Version = incoming.Version

	if err := h.orm.WithContext(ctx).Save(&existing).Error(); err != nil {
		return fmt.Errorf("updating maintenance execution: %w", err)
	}
	return nil
}

func (h *MaintenanceExecutionHandler) extractMaintenanceExecutionFields(message pubsub.Message) MaintenanceExecutionData {
	avroExecution, ok := message.(*avro.AvroMaintenanceExecution)
	if !ok {
		slog.Error("message is not *avro.AvroMaintenanceExecution", "message", message)
		return MaintenanceExecutionData{}
	}

	fieldValuesBytes, _ := json.Marshal(avroExecution.FieldValues)
	fieldValues := string(fieldValuesBytes)

	var completedAt *utils.Time
	if avroExecution.CompletedAt != nil {
		completedAt = &utils.Time{Time: *avroExecution.CompletedAt}
	}

	var deletedAt *utils.Time
	if avroExecution.DeletedAt != nil {
		deletedAt = &utils.Time{Time: *avroExecution.DeletedAt}
	}

	return MaintenanceExecutionData{
		ID:            avroExecution.ID,
		Version:       avroExecution.Version,
		ActivityID:    avroExecution.ActivityID,
		ScheduledDate: utils.Time{Time: avroExecution.ScheduledDate},
		CompletedAt:   completedAt,
		CompletedBy:   avroExecution.CompletedBy,
		OverdueDays:   avroExecution.OverdueDays,
		FieldValues:   fieldValues,
		CreatedAt:     utils.Time{Time: avroExecution.CreatedAt},
		UpdatedAt:     utils.Time{Time: avroExecution.UpdatedAt},
		DeletedAt:     deletedAt,
	}
}

func (h *MaintenanceExecutionHandler) toDomainMaintenanceExecution(internalExecution MaintenanceExecutionData) map[string]any {
	return map[string]any{
		"id":          internalExecution.ID,
		"activity_id": internalExecution.ActivityID,
	}
}
