package persistence

import (
	"testing"
	"time"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/shared_kernel/domain"
)

func TestTenantConfigurationRepository_Create(t *testing.T) {
	// This is a basic test to verify the repository compiles and follows the pattern
	// In a real implementation, you would use a test database and mock the publisher

	config := domain.TenantConfiguration{
		ID:        "test-id",
		TenantID:  "tenant-id",
		Timezone:  "UTC",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test the internal model conversion
	internalConfig := internal.FromTenantConfiguration(config)
	domainConfig := internalConfig.ToDomain()

	if domainConfig.ID != config.ID {
		t.Errorf("expected ID %s, got %s", config.ID, domainConfig.ID)
	}

	if domainConfig.TenantID != config.TenantID {
		t.Errorf("expected TenantID %s, got %s", config.TenantID, domainConfig.TenantID)
	}

	if domainConfig.Timezone != config.Timezone {
		t.Errorf("expected Timezone %s, got %s", config.Timezone, domainConfig.Timezone)
	}
}

func TestTenantConfigurationBuilder(t *testing.T) {
	tenantID := domain.ID("test-tenant-id")
	timezone := "America/New_York"

	config, err := domain.NewTenantConfigurationBuilder().
		WithTenantID(tenantID).
		WithTimezone(timezone).
		Build()

	if err != nil {
		t.Fatalf("failed to build tenant configuration: %v", err)
	}

	if config.TenantID != tenantID {
		t.Errorf("expected TenantID %s, got %s", tenantID, config.TenantID)
	}

	if config.Timezone != timezone {
		t.Errorf("expected Timezone %s, got %s", timezone, config.Timezone)
	}

	if config.ID == "" {
		t.Error("expected ID to be generated")
	}

	if config.Version != 1 {
		t.Errorf("expected Version 1, got %d", config.Version)
	}
}
