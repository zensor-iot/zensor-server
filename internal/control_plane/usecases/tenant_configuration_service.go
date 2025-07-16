package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"zensor-server/internal/shared_kernel/domain"
)

var (
	ErrInvalidTimezone = errors.New("invalid timezone")
)

type TenantConfigurationService interface {
	CreateTenantConfiguration(ctx context.Context, config domain.TenantConfiguration) error
	GetTenantConfiguration(ctx context.Context, tenantID domain.ID) (domain.TenantConfiguration, error)
	UpdateTenantConfiguration(ctx context.Context, config domain.TenantConfiguration) error
	GetOrCreateTenantConfiguration(ctx context.Context, tenantID domain.ID, defaultTimezone string) (domain.TenantConfiguration, error)
}

func NewTenantConfigurationService(repository TenantConfigurationRepository) *SimpleTenantConfigurationService {
	return &SimpleTenantConfigurationService{
		repository: repository,
	}
}

var _ TenantConfigurationService = &SimpleTenantConfigurationService{}

type SimpleTenantConfigurationService struct {
	repository TenantConfigurationRepository
}

func (s *SimpleTenantConfigurationService) CreateTenantConfiguration(ctx context.Context, config domain.TenantConfiguration) error {
	// Check if configuration already exists for this tenant
	existingConfig, err := s.repository.GetByTenantID(ctx, config.TenantID)
	if err != nil && !errors.Is(err, ErrTenantConfigurationNotFound) {
		slog.Error("checking existing tenant configuration", slog.String("error", err.Error()))
		return fmt.Errorf("checking existing tenant configuration: %w", err)
	}

	if existingConfig.ID != "" {
		slog.Warn("tenant configuration already exists", slog.String("tenant_id", config.TenantID.String()))
		return fmt.Errorf("tenant configuration already exists for tenant %s", config.TenantID.String())
	}

	err = s.repository.Create(ctx, config)
	if err != nil {
		slog.Error("creating tenant configuration", slog.String("error", err.Error()))
		return fmt.Errorf("creating tenant configuration: %w", err)
	}

	slog.Info("tenant configuration created successfully",
		slog.String("id", config.ID.String()),
		slog.String("tenant_id", config.TenantID.String()))

	return nil
}

func (s *SimpleTenantConfigurationService) GetTenantConfiguration(ctx context.Context, tenantID domain.ID) (domain.TenantConfiguration, error) {
	config, err := s.repository.GetByTenantID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, ErrTenantConfigurationNotFound) {
			return domain.TenantConfiguration{}, ErrTenantConfigurationNotFound
		}
		slog.Error("getting tenant configuration", slog.String("error", err.Error()))
		return domain.TenantConfiguration{}, fmt.Errorf("getting tenant configuration: %w", err)
	}

	return config, nil
}

func (s *SimpleTenantConfigurationService) UpdateTenantConfiguration(ctx context.Context, config domain.TenantConfiguration) error {
	existingConfig, err := s.repository.GetByTenantID(ctx, config.TenantID)
	if err != nil {
		if errors.Is(err, ErrTenantConfigurationNotFound) {
			return ErrTenantConfigurationNotFound
		}
		return fmt.Errorf("getting tenant configuration: %w", err)
	}

	// Update the existing configuration (version handled internally by the domain model)
	err = existingConfig.UpdateTimezone(config.Timezone)
	if err != nil {
		slog.Error("updating tenant configuration timezone", slog.String("error", err.Error()))
		return ErrInvalidTimezone
	}

	err = s.repository.Update(ctx, existingConfig)
	if err != nil {
		slog.Error("updating tenant configuration", slog.String("error", err.Error()))
		return fmt.Errorf("updating tenant configuration: %w", err)
	}

	slog.Info("tenant configuration updated successfully",
		slog.String("tenant_id", config.TenantID.String()),
		slog.Int("version", existingConfig.Version))
	return nil
}

func (s *SimpleTenantConfigurationService) GetOrCreateTenantConfiguration(ctx context.Context, tenantID domain.ID, defaultTimezone string) (domain.TenantConfiguration, error) {
	config, err := s.repository.GetByTenantID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, ErrTenantConfigurationNotFound) {
			// Create default configuration
			newConfig, err := domain.NewTenantConfigurationBuilder().
				WithTenantID(tenantID).
				WithTimezone(defaultTimezone).
				Build()
			if err != nil {
				slog.Error("building default tenant configuration", slog.String("error", err.Error()))
				return domain.TenantConfiguration{}, fmt.Errorf("building default tenant configuration: %w", err)
			}

			err = s.repository.Create(ctx, newConfig)
			if err != nil {
				slog.Error("creating default tenant configuration", slog.String("error", err.Error()))
				return domain.TenantConfiguration{}, fmt.Errorf("creating default tenant configuration: %w", err)
			}

			slog.Info("default tenant configuration created",
				slog.String("tenant_id", tenantID.String()),
				slog.String("timezone", defaultTimezone))

			return newConfig, nil
		}
		slog.Error("getting tenant configuration", slog.String("error", err.Error()))
		return domain.TenantConfiguration{}, fmt.Errorf("getting tenant configuration: %w", err)
	}

	return config, nil
}
