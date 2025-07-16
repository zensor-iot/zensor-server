package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
)

// TenantConfigurationData represents the tenant_configurations table structure for GORM operations
type TenantConfigurationData struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	TenantID  string    `json:"tenant_id" gorm:"uniqueIndex;not null"`
	Timezone  string    `json:"timezone" gorm:"not null"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TenantConfigurationData) TableName() string {
	return "tenant_configurations"
}

// TenantConfigurationHandler handles replication of tenant configuration data
type TenantConfigurationHandler struct {
	orm sql.ORM
}

// NewTenantConfigurationHandler creates a new tenant configuration handler
func NewTenantConfigurationHandler(orm sql.ORM) *TenantConfigurationHandler {
	return &TenantConfigurationHandler{
		orm: orm,
	}
}

// TopicName returns the topic name for tenant configurations
func (h *TenantConfigurationHandler) TopicName() pubsub.Topic {
	return "tenant_configurations"
}

// GetByID retrieves a tenant configuration by ID
func (h *TenantConfigurationHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var entity TenantConfigurationData
	err := h.orm.WithContext(ctx).First(&entity, "id = ?", id).Error()
	if err != nil {
		if errors.Is(err, sql.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant configuration not found")
		}
		return nil, err
	}

	// Convert to Avro format for consistency with other handlers
	avroConfig := &avro.AvroTenantConfiguration{
		ID:        entity.ID,
		TenantID:  entity.TenantID,
		Timezone:  entity.Timezone,
		Version:   entity.Version,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}

	return avroConfig, nil
}

// Create creates a new tenant configuration
func (h *TenantConfigurationHandler) Create(ctx context.Context, key pubsub.Key, msg pubsub.Message) error {
	internalConfig := h.extractTenantConfigurationFields(msg)

	err := h.orm.WithContext(ctx).Create(&internalConfig).Error()
	if err != nil {
		return fmt.Errorf("creating tenant configuration: %w", err)
	}

	return nil
}

// Update updates an existing tenant configuration
func (h *TenantConfigurationHandler) Update(ctx context.Context, key pubsub.Key, msg pubsub.Message) error {
	var existing TenantConfigurationData
	if err := h.orm.WithContext(ctx).First(&existing, "id = ?", string(key)).Error(); err != nil {
		return fmt.Errorf("fetching existing tenant configuration: %w", err)
	}

	incoming := h.extractTenantConfigurationFields(msg)
	existing.TenantID = incoming.TenantID
	existing.Timezone = incoming.Timezone
	existing.Version = incoming.Version
	existing.UpdatedAt = incoming.UpdatedAt

	if err := h.orm.WithContext(ctx).Save(&existing).Error(); err != nil {
		return fmt.Errorf("updating tenant configuration: %w", err)
	}

	return nil
}

// extractTenantConfigurationFields uses reflection to extract tenant configuration fields from any message type
func (h *TenantConfigurationHandler) extractTenantConfigurationFields(message pubsub.Message) TenantConfigurationData {
	avroConfig, ok := message.(*avro.AvroTenantConfiguration)
	if !ok {
		slog.Error("message is not *avro.AvroTenantConfiguration", "message", message)
		return TenantConfigurationData{}
	}

	return TenantConfigurationData{
		ID:        avroConfig.ID,
		TenantID:  avroConfig.TenantID,
		Timezone:  avroConfig.Timezone,
		Version:   int(avroConfig.Version),
		CreatedAt: avroConfig.CreatedAt,
		UpdatedAt: avroConfig.UpdatedAt,
	}
}
