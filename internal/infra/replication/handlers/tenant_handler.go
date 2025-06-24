package handlers

import (
	"context"
	"fmt"
	"reflect"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
)

// TenantGORM represents the tenant table structure for GORM operations
type TenantGORM struct {
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

func (TenantGORM) TableName() string {
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
	var internalTenant TenantGORM

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
func (h *TenantHandler) extractTenantFields(message pubsub.Message) TenantGORM {
	val := reflect.ValueOf(message)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := TenantGORM{
		Version: 1,
	}

	// Extract ID field
	if idField := val.FieldByName("ID"); idField.IsValid() {
		if idField.Type().String() == "domain.ID" {
			// Handle domain.ID type by calling String() method
			if stringMethod := idField.MethodByName("String"); stringMethod.IsValid() {
				results := stringMethod.Call(nil)
				if len(results) > 0 {
					result.ID = results[0].String()
				}
			}
		} else {
			result.ID = idField.Interface().(string)
		}
	}

	// Extract Name field
	if nameField := val.FieldByName("Name"); nameField.IsValid() {
		result.Name = nameField.Interface().(string)
	}

	// Extract Email field
	if emailField := val.FieldByName("Email"); emailField.IsValid() {
		result.Email = emailField.Interface().(string)
	}

	// Extract Description field
	if descField := val.FieldByName("Description"); descField.IsValid() {
		result.Description = descField.Interface().(string)
	}

	// Extract IsActive field
	if activeField := val.FieldByName("IsActive"); activeField.IsValid() {
		result.IsActive = activeField.Interface().(bool)
	}

	// Extract CreatedAt field
	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		result.CreatedAt = createdAtField.Interface().(time.Time)
	}

	// Extract UpdatedAt field
	if updatedAtField := val.FieldByName("UpdatedAt"); updatedAtField.IsValid() {
		result.UpdatedAt = updatedAtField.Interface().(time.Time)
	}

	// Extract DeletedAt field
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

// toDomainTenant converts internal tenant to domain representation
func (h *TenantHandler) toDomainTenant(internalTenant TenantGORM) map[string]any {
	return map[string]any{
		"id":   internalTenant.ID,
		"name": internalTenant.Name,
	}
}
