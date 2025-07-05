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

// TenantData represents the tenant table structure for GORM operations
type TenantData struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Version     int        `json:"version"`
	Name        string     `json:"name" gorm:"uniqueIndex;not null"`
	Email       string     `json:"email" gorm:"not null"`
	Description string     `json:"description"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (TenantData) TableName() string {
	return "tenants"
}

// TenantHandler handles replication of tenant data
type TenantHandler struct {
	orm sql.ORM
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(orm sql.ORM) *TenantHandler {
	return &TenantHandler{
		orm: orm,
	}
}

// TopicName returns the tenants topic
func (h *TenantHandler) TopicName() pubsub.Topic {
	return "tenants"
}

// Create handles creating a new tenant record
func (h *TenantHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	internalTenant := h.extractTenantFields(message)

	err := h.orm.WithContext(ctx).Create(&internalTenant).Error()
	if err != nil {
		return fmt.Errorf("creating tenant: %w", err)
	}

	return nil
}

// GetByID retrieves a tenant by its ID
func (h *TenantHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var internalTenant TenantData

	err := h.orm.WithContext(ctx).First(&internalTenant, "id = ?", id).Error()
	if err != nil {
		return nil, fmt.Errorf("getting tenant: %w", err)
	}

	tenant := h.toDomainTenant(internalTenant)
	return tenant, nil
}

// Update handles updating an existing tenant record
func (h *TenantHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	var existing TenantData
	if err := h.orm.WithContext(ctx).First(&existing, "id = ?", string(key)).Error(); err != nil {
		return fmt.Errorf("fetching existing tenant: %w", err)
	}
	incoming := h.extractTenantFields(message)
	if incoming.Name != "" {
		existing.Name = incoming.Name
	}
	if incoming.Email != "" {
		existing.Email = incoming.Email
	}
	existing.Description = incoming.Description
	existing.IsActive = incoming.IsActive
	existing.UpdatedAt = incoming.UpdatedAt
	existing.DeletedAt = incoming.DeletedAt
	existing.Version = incoming.Version

	if err := h.orm.WithContext(ctx).Save(&existing).Error(); err != nil {
		return fmt.Errorf("updating tenant: %w", err)
	}
	return nil
}

// extractTenantFields uses reflection to extract tenant fields from any message type
func (h *TenantHandler) extractTenantFields(message pubsub.Message) TenantData {
	avroTenant, ok := message.(*avro.AvroTenant)
	if !ok {
		slog.Error("message is not *avro.AvroTenant", "message", message)
		return TenantData{}
	}

	return TenantData{
		ID:          avroTenant.ID,
		Version:     int(avroTenant.Version),
		Name:        avroTenant.Name,
		Email:       avroTenant.Email,
		Description: avroTenant.Description,
		IsActive:    avroTenant.IsActive,
		CreatedAt:   avroTenant.CreatedAt,
		UpdatedAt:   avroTenant.UpdatedAt,
		DeletedAt:   avroTenant.DeletedAt,
	}
}

func (h *TenantHandler) toDomainTenant(internalTenant TenantData) map[string]any {
	return map[string]any{
		"id":   internalTenant.ID,
		"name": internalTenant.Name,
	}
}
