package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
)

type Tenant struct {
	ID          ID
	Name        string
	Email       string
	Description string
	IsActive    bool
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // For soft deletion
}

func (t *Tenant) IsDeleted() bool {
	return t.DeletedAt != nil
}

func (t *Tenant) SoftDelete() {
	now := time.Now()
	t.DeletedAt = &now
	t.IsActive = false
	t.UpdatedAt = now
}

func (t *Tenant) Activate() {
	t.IsActive = true
	t.UpdatedAt = time.Now()
}

func (t *Tenant) Deactivate() {
	t.IsActive = false
	t.UpdatedAt = time.Now()
}

func (t *Tenant) UpdateInfo(name, email, description string) {
	if name != "" {
		t.Name = name
	}
	if email != "" {
		t.Email = email
	}
	if description != "" {
		t.Description = description
	}
	t.UpdatedAt = time.Now()
}

func NewTenantBuilder() *tenantBuilder {
	return &tenantBuilder{}
}

type tenantBuilder struct {
	actions []tenantHandler
}

type tenantHandler func(t *Tenant) error

func (b *tenantBuilder) WithName(name string) *tenantBuilder {
	b.actions = append(b.actions, func(t *Tenant) error {
		t.Name = name
		return nil
	})
	return b
}

func (b *tenantBuilder) WithEmail(email string) *tenantBuilder {
	b.actions = append(b.actions, func(t *Tenant) error {
		t.Email = email
		return nil
	})
	return b
}

func (b *tenantBuilder) WithDescription(description string) *tenantBuilder {
	b.actions = append(b.actions, func(t *Tenant) error {
		t.Description = description
		return nil
	})
	return b
}

func (b *tenantBuilder) Build() (Tenant, error) {
	now := time.Now()
	result := Tenant{
		ID:        ID(utils.GenerateUUID()),
		IsActive:  true,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, action := range b.actions {
		if err := action(&result); err != nil {
			return Tenant{}, err
		}
	}

	return result, nil
}
