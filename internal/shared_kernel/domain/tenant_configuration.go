package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
)

type TenantConfiguration struct {
	ID        ID
	TenantID  ID
	Timezone  string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tc *TenantConfiguration) UpdateTimezone(timezone string) error {
	if err := utils.ValidateTimezone(timezone); err != nil {
		return err
	}
	tc.Timezone = timezone
	tc.Version++
	tc.UpdatedAt = time.Now()
	return nil
}

func NewTenantConfigurationBuilder() *tenantConfigurationBuilder {
	return &tenantConfigurationBuilder{}
}

type tenantConfigurationBuilder struct {
	actions []tenantConfigurationHandler
}

type tenantConfigurationHandler func(tc *TenantConfiguration) error

func (b *tenantConfigurationBuilder) WithTenantID(tenantID ID) *tenantConfigurationBuilder {
	b.actions = append(b.actions, func(tc *TenantConfiguration) error {
		tc.TenantID = tenantID
		return nil
	})
	return b
}

func (b *tenantConfigurationBuilder) WithTimezone(timezone string) *tenantConfigurationBuilder {
	b.actions = append(b.actions, func(tc *TenantConfiguration) error {
		if err := utils.ValidateTimezone(timezone); err != nil {
			return err
		}
		tc.Timezone = timezone
		return nil
	})
	return b
}

func (b *tenantConfigurationBuilder) Build() (TenantConfiguration, error) {
	now := time.Now()
	result := TenantConfiguration{
		ID:        ID(utils.GenerateUUID()),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, action := range b.actions {
		if err := action(&result); err != nil {
			return TenantConfiguration{}, err
		}
	}

	return result, nil
}
