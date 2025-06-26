package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"zensor-server/internal/control_plane/domain"
)

var (
	ErrTenantNotFound        = errors.New("tenant not found")
	ErrTenantDuplicated      = errors.New("tenant already exists")
	ErrTenantSoftDeleted     = errors.New("tenant is soft deleted")
	ErrTenantVersionConflict = errors.New("tenant version conflict")
)

type TenantService interface {
	CreateTenant(ctx context.Context, tenant domain.Tenant) error
	GetTenant(ctx context.Context, id domain.ID) (domain.Tenant, error)
	GetTenantByName(ctx context.Context, name string) (domain.Tenant, error)
	ListTenants(ctx context.Context, includeDeleted bool, pagination Pagination) ([]domain.Tenant, int, error)
	UpdateTenant(ctx context.Context, tenant domain.Tenant) error
	SoftDeleteTenant(ctx context.Context, id domain.ID) error
	ActivateTenant(ctx context.Context, id domain.ID) error
	DeactivateTenant(ctx context.Context, id domain.ID) error
	AdoptDevice(ctx context.Context, tenantID, deviceID domain.ID) error
	ListTenantDevices(ctx context.Context, tenantID domain.ID, pagination Pagination) ([]domain.Device, int, error)
}

type TenantRepository interface {
	Create(ctx context.Context, tenant domain.Tenant) error
	GetByID(ctx context.Context, id domain.ID) (domain.Tenant, error)
	GetByName(ctx context.Context, name string) (domain.Tenant, error)
	Update(ctx context.Context, tenant domain.Tenant) error
	FindAll(ctx context.Context, includeDeleted bool, pagination Pagination) ([]domain.Tenant, int, error)
}

func NewTenantService(repository TenantRepository, deviceService DeviceService) *SimpleTenantService {
	return &SimpleTenantService{
		repository:    repository,
		deviceService: deviceService,
	}
}

var _ TenantService = &SimpleTenantService{}

type SimpleTenantService struct {
	repository    TenantRepository
	deviceService DeviceService
}

func (s *SimpleTenantService) CreateTenant(ctx context.Context, tenant domain.Tenant) error {
	existingTenant, err := s.repository.GetByName(ctx, tenant.Name)
	if err != nil && !errors.Is(err, ErrTenantNotFound) {
		slog.Error("checking existing tenant", slog.String("error", err.Error()))
		return fmt.Errorf("checking existing tenant: %w", err)
	}

	if existingTenant.ID != "" {
		slog.Warn("tenant already exists", slog.String("name", tenant.Name))
		return ErrTenantDuplicated
	}

	err = s.repository.Create(ctx, tenant)
	if err != nil {
		slog.Error("creating tenant", slog.String("error", err.Error()))
		return fmt.Errorf("creating tenant: %w", err)
	}

	slog.Info("tenant created successfully",
		slog.String("id", tenant.ID.String()),
		slog.String("name", tenant.Name))

	return nil
}

func (s *SimpleTenantService) GetTenant(ctx context.Context, id domain.ID) (domain.Tenant, error) {
	tenant, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return domain.Tenant{}, ErrTenantNotFound
		}
		slog.Error("getting tenant", slog.String("error", err.Error()))
		return domain.Tenant{}, fmt.Errorf("getting tenant: %w", err)
	}

	return tenant, nil
}

func (s *SimpleTenantService) GetTenantByName(ctx context.Context, name string) (domain.Tenant, error) {
	tenant, err := s.repository.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return domain.Tenant{}, ErrTenantNotFound
		}
		slog.Error("getting tenant by name", slog.String("error", err.Error()))
		return domain.Tenant{}, fmt.Errorf("getting tenant by name: %w", err)
	}

	return tenant, nil
}

func (s *SimpleTenantService) ListTenants(ctx context.Context, includeDeleted bool, pagination Pagination) ([]domain.Tenant, int, error) {
	tenants, total, err := s.repository.FindAll(ctx, includeDeleted, pagination)
	if err != nil {
		slog.Error("listing tenants", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("listing tenants: %w", err)
	}

	return tenants, total, nil
}

func (s *SimpleTenantService) UpdateTenant(ctx context.Context, tenant domain.Tenant) error {
	existingTenant, err := s.repository.GetByID(ctx, tenant.ID)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return ErrTenantNotFound
		}
		return fmt.Errorf("getting tenant: %w", err)
	}

	if existingTenant.IsDeleted() {
		return ErrTenantSoftDeleted
	}

	// Check version for optimistic locking
	if tenant.Version != 0 && tenant.Version != existingTenant.Version {
		return ErrTenantVersionConflict
	}

	// Check if new name conflicts with existing tenant
	if tenant.Name != "" && tenant.Name != existingTenant.Name {
		existing, err := s.repository.GetByName(ctx, tenant.Name)
		if err != nil && !errors.Is(err, ErrTenantNotFound) {
			return fmt.Errorf("checking name conflict: %w", err)
		}
		if err == nil && existing.ID != tenant.ID {
			return ErrTenantDuplicated
		}
	}

	existingTenant.UpdateInfo(tenant.Name, tenant.Email, tenant.Description)

	err = s.repository.Update(ctx, existingTenant)
	if err != nil {
		slog.Error("updating tenant", slog.String("error", err.Error()))
		return fmt.Errorf("updating tenant: %w", err)
	}

	slog.Info("tenant updated successfully", slog.String("id", tenant.ID.String()), slog.Int("version", existingTenant.Version))
	return nil
}

func (s *SimpleTenantService) SoftDeleteTenant(ctx context.Context, id domain.ID) error {
	tenant, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return ErrTenantNotFound
		}
		return fmt.Errorf("getting tenant: %w", err)
	}

	if tenant.IsDeleted() {
		return ErrTenantSoftDeleted
	}

	tenant.SoftDelete()

	err = s.repository.Update(ctx, tenant)
	if err != nil {
		slog.Error("soft deleting tenant", slog.String("error", err.Error()))
		return fmt.Errorf("soft deleting tenant: %w", err)
	}

	slog.Info("tenant soft deleted successfully", slog.String("id", id.String()))
	return nil
}

func (s *SimpleTenantService) ActivateTenant(ctx context.Context, id domain.ID) error {
	tenant, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return ErrTenantNotFound
		}
		return fmt.Errorf("getting tenant: %w", err)
	}

	if tenant.IsDeleted() {
		return ErrTenantSoftDeleted
	}

	tenant.Activate()

	err = s.repository.Update(ctx, tenant)
	if err != nil {
		slog.Error("activating tenant", slog.String("error", err.Error()))
		return fmt.Errorf("activating tenant: %w", err)
	}

	slog.Info("tenant activated successfully", slog.String("id", id.String()))
	return nil
}

func (s *SimpleTenantService) DeactivateTenant(ctx context.Context, id domain.ID) error {
	tenant, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return ErrTenantNotFound
		}
		return fmt.Errorf("getting tenant: %w", err)
	}

	if tenant.IsDeleted() {
		return ErrTenantSoftDeleted
	}

	tenant.Deactivate()

	err = s.repository.Update(ctx, tenant)
	if err != nil {
		slog.Error("deactivating tenant", slog.String("error", err.Error()))
		return fmt.Errorf("deactivating tenant: %w", err)
	}

	slog.Info("tenant deactivated successfully", slog.String("id", id.String()))
	return nil
}

func (s *SimpleTenantService) AdoptDevice(ctx context.Context, tenantID, deviceID domain.ID) error {
	// First verify that the tenant exists and is not soft deleted
	tenant, err := s.repository.GetByID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return ErrTenantNotFound
		}
		return fmt.Errorf("getting tenant: %w", err)
	}

	if tenant.IsDeleted() {
		return ErrTenantSoftDeleted
	}

	// Delegate to the device service to handle the adoption
	err = s.deviceService.AdoptDeviceToTenant(ctx, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("adopting device to tenant: %w", err)
	}

	slog.Info("device adopted to tenant successfully",
		slog.String("tenant_id", tenantID.String()),
		slog.String("device_id", deviceID.String()))

	return nil
}

func (s *SimpleTenantService) ListTenantDevices(ctx context.Context, tenantID domain.ID, pagination Pagination) ([]domain.Device, int, error) {
	// First verify that the tenant exists and is not soft deleted
	tenant, err := s.repository.GetByID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return nil, 0, ErrTenantNotFound
		}
		return nil, 0, fmt.Errorf("getting tenant: %w", err)
	}

	if tenant.IsDeleted() {
		return nil, 0, ErrTenantSoftDeleted
	}

	// Get devices belonging to this tenant
	devices, total, err := s.deviceService.DevicesByTenant(ctx, tenantID, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("getting devices for tenant: %w", err)
	}

	slog.Info("retrieved devices for tenant",
		slog.String("tenant_id", tenantID.String()),
		slog.Int("device_count", len(devices)))

	return devices, total, nil
}
