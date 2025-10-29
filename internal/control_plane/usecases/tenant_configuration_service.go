package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"zensor-server/internal/shared_kernel/domain"
)

var (
	ErrInvalidTimezone                    = errors.New("invalid timezone")
	ErrForbiddenTenantConfigurationAccess = errors.New("forbidden tenant configuration access")
)

func NewTenantConfigurationService(repository TenantConfigurationRepository, userService UserService) *SimpleTenantConfigurationService {
	return &SimpleTenantConfigurationService{
		repository:  repository,
		userService: userService,
	}
}

var _ TenantConfigurationService = &SimpleTenantConfigurationService{}

type SimpleTenantConfigurationService struct {
	repository  TenantConfigurationRepository
	userService UserService
}

func (s *SimpleTenantConfigurationService) UpsertTenantConfiguration(ctx context.Context, userEmail string, config domain.TenantConfiguration) (domain.TenantConfiguration, error) {
	// Convert email to domain.ID for user lookup
	userID := domain.ID(userEmail)
	user, err := s.userService.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			slog.Warn("user not found", slog.String("user_email", userEmail))
			return domain.TenantConfiguration{}, ErrUserNotFound
		}
		return domain.TenantConfiguration{}, fmt.Errorf("getting user: %w", err)
	}

	if !user.HasTenant(config.TenantID) {
		slog.Warn("user does not have permission to access tenant configuration",
			slog.String("user_email", userEmail),
			slog.String("tenant_id", config.TenantID.String()))
		return domain.TenantConfiguration{}, ErrForbiddenTenantConfigurationAccess
	}

	existingConfig, err := s.repository.GetByTenantID(ctx, config.TenantID)
	if err != nil {
		if errors.Is(err, ErrTenantConfigurationNotFound) {
			err = s.repository.Create(ctx, config)
			if err != nil {
				slog.Error("creating tenant configuration", slog.String("error", err.Error()))
				return domain.TenantConfiguration{}, fmt.Errorf("creating tenant configuration: %w", err)
			}

			slog.Info("tenant configuration created successfully",
				slog.String("id", config.ID.String()),
				slog.String("tenant_id", config.TenantID.String()))

			return config, nil
		}
		slog.Error("getting tenant configuration", slog.String("error", err.Error()))
		return domain.TenantConfiguration{}, fmt.Errorf("getting tenant configuration: %w", err)
	}

	err = existingConfig.UpdateTimezone(config.Timezone)
	if err != nil {
		slog.Error("updating tenant configuration timezone", slog.String("error", err.Error()))
		return domain.TenantConfiguration{}, ErrInvalidTimezone
	}

	if config.NotificationEmail != "" {
		existingConfig.NotificationEmail = config.NotificationEmail
	}

	err = s.repository.Update(ctx, existingConfig)
	if err != nil {
		slog.Error("updating tenant configuration", slog.String("error", err.Error()))
		return domain.TenantConfiguration{}, fmt.Errorf("updating tenant configuration: %w", err)
	}

	slog.Info("tenant configuration updated successfully",
		slog.String("tenant_id", config.TenantID.String()),
		slog.Int("version", existingConfig.Version))

	return existingConfig, nil
}

func (s *SimpleTenantConfigurationService) GetTenantConfiguration(ctx context.Context, tenant domain.Tenant) (domain.TenantConfiguration, error) {
	config, err := s.repository.GetByTenantID(ctx, tenant.ID)
	if err != nil {
		if errors.Is(err, ErrTenantConfigurationNotFound) {
			return domain.TenantConfiguration{}, ErrTenantConfigurationNotFound
		}
		slog.Error("getting tenant configuration", slog.String("error", err.Error()))
		return domain.TenantConfiguration{}, fmt.Errorf("getting tenant configuration: %w", err)
	}

	return config, nil
}

func (s *SimpleTenantConfigurationService) GetOrCreateTenantConfiguration(ctx context.Context, tenant domain.Tenant, defaultTimezone string) (domain.TenantConfiguration, error) {
	config, err := s.repository.GetByTenantID(ctx, tenant.ID)
	if err != nil {
		if errors.Is(err, ErrTenantConfigurationNotFound) {
			newConfig, err := domain.NewTenantConfigurationBuilder().
				WithTenantID(tenant.ID).
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
				slog.String("tenant_id", tenant.ID.String()),
				slog.String("timezone", defaultTimezone))

			return newConfig, nil
		}
		slog.Error("getting tenant configuration", slog.String("error", err.Error()))
		return domain.TenantConfiguration{}, fmt.Errorf("getting tenant configuration: %w", err)
	}

	return config, nil
}
