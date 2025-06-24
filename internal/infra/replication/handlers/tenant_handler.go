package handlers

import (
	"context"
	"fmt"
	"reflect"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
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
	return "tenants_final"
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
	internalTenant := h.extractTenantFields(message)

	err := h.orm.WithContext(ctx).Save(&internalTenant).Error()
	if err != nil {
		return fmt.Errorf("updating tenant: %w", err)
	}

	return nil
}

// extractTenantFields uses reflection to extract tenant fields from any message type
func (h *TenantHandler) extractTenantFields(message pubsub.Message) TenantData {
	val := reflect.ValueOf(message)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := TenantData{
		Version: 1,
	}

	if idField := val.FieldByName("ID"); idField.IsValid() {
		result.ID = idField.Interface().(string)
	}

	if nameField := val.FieldByName("Name"); nameField.IsValid() {
		result.Name = nameField.Interface().(string)
	}

	if emailField := val.FieldByName("Email"); emailField.IsValid() {
		result.Email = emailField.Interface().(string)
	}

	if descField := val.FieldByName("Description"); descField.IsValid() {
		result.Description = descField.Interface().(string)
	}

	if activeField := val.FieldByName("IsActive"); activeField.IsValid() {
		result.IsActive = activeField.Interface().(bool)
	}

	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		result.CreatedAt = createdAtField.Interface().(time.Time)
	}

	if updatedAtField := val.FieldByName("UpdatedAt"); updatedAtField.IsValid() {
		result.UpdatedAt = updatedAtField.Interface().(time.Time)
	}

	if deletedAtField := val.FieldByName("DeletedAt"); deletedAtField.IsValid() {
		if deletedAtField.IsNil() {
			result.DeletedAt = nil
		} else {
			deletedAt := deletedAtField.Interface().(*time.Time)
			result.DeletedAt = deletedAt
		}
	}

	return result
}

func (h *TenantHandler) toDomainTenant(internalTenant TenantData) map[string]any {
	return map[string]any{
		"id":   internalTenant.ID,
		"name": internalTenant.Name,
	}
}
