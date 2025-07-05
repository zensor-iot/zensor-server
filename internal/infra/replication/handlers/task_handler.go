package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
)

// TaskData represents the task table structure for GORM operations
type TaskData struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Version   int       `json:"version"`
	DeviceID  string    `json:"device_id" gorm:"foreignKey:device_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TaskData) TableName() string {
	return "tasks_final"
}

// TaskHandler handles replication of task data
type TaskHandler struct {
	orm sql.ORM
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(orm sql.ORM) *TaskHandler {
	return &TaskHandler{
		orm: orm,
	}
}

// TopicName returns the tasks topic
func (h *TaskHandler) TopicName() pubsub.Topic {
	return "tasks"
}

// Create handles creating a new task record
func (h *TaskHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	internalTask := h.extractTaskFields(message)

	err := h.orm.WithContext(ctx).Create(&internalTask).Error()
	if err != nil {
		return fmt.Errorf("creating task: %w", err)
	}

	return nil
}

// GetByID retrieves a task by its ID
func (h *TaskHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var internalTask TaskData

	err := h.orm.WithContext(ctx).First(&internalTask, "id = ?", id).Error()
	if err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}

	task := h.toDomainTask(internalTask)
	return task, nil
}

// Update handles updating an existing task record
func (h *TaskHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	internalTask := h.extractTaskFields(message)

	err := h.orm.WithContext(ctx).Save(&internalTask).Error()
	if err != nil {
		return fmt.Errorf("updating task: %w", err)
	}

	return nil
}

// extractTaskFields uses reflection to extract task fields from any message type
func (h *TaskHandler) extractTaskFields(message pubsub.Message) TaskData {
	avroTask, ok := message.(*avro.AvroTask)
	if !ok {
		slog.Error("message is not *avro.AvroTask", "message", message)
		return TaskData{}
	}

	return TaskData{
		ID:        avroTask.ID,
		Version:   int(avroTask.Version),
		DeviceID:  avroTask.DeviceID,
		CreatedAt: avroTask.CreatedAt,
		UpdatedAt: avroTask.UpdatedAt,
	}
}

func (h *TaskHandler) toDomainTask(internalTask TaskData) map[string]any {
	return map[string]any{
		"id":         internalTask.ID,
		"version":    internalTask.Version,
		"device_id":  internalTask.DeviceID,
		"created_at": internalTask.CreatedAt,
		"updated_at": internalTask.UpdatedAt,
	}
}
