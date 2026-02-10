package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"zensor-server/internal/shared_kernel/domain"
)

func NewUserService(
	repository UserRepository,
	tenantRepository TenantRepository,
) *SimpleUserService {
	return &SimpleUserService{
		repository:       repository,
		tenantRepository: tenantRepository,
	}
}

var _ UserService = &SimpleUserService{}

type SimpleUserService struct {
	repository       UserRepository
	tenantRepository TenantRepository
}

func (s *SimpleUserService) AssociateTenants(ctx context.Context, userID domain.ID, tenantIDs []domain.ID) error {
	var invalidTenants []domain.ID
	var validTenants []domain.ID

	for _, tenantID := range tenantIDs {
		_, err := s.tenantRepository.GetByID(ctx, tenantID)
		if errors.Is(err, ErrTenantNotFound) {
			slog.Warn("tenant not found", slog.String("tenant_id", tenantID.String()))
			invalidTenants = append(invalidTenants, tenantID)
		} else if err != nil {
			slog.Error("getting tenant", slog.String("error", err.Error()))
			return fmt.Errorf("validating tenant: %w", err)
		} else {
			validTenants = append(validTenants, tenantID)
		}
	}

	// Determine the appropriate error based on validation results
	if len(invalidTenants) > 0 {
		if len(validTenants) > 0 {
			// Mixed valid and invalid tenants
			return ErrMixedTenantValidation
		} else {
			// All tenants are invalid
			return ErrTenantNotFound
		}
	}

	// All tenants are valid, proceed with association
	user := domain.User{
		ID:      userID,
		Tenants: tenantIDs,
	}

	err := s.repository.Upsert(ctx, user)
	if err != nil {
		slog.Error("associating tenants with user", slog.String("error", err.Error()))
		return fmt.Errorf("associating tenants: %w", err)
	}

	slog.Info("tenants associated with user",
		slog.String("user_id", userID.String()),
		slog.Int("tenant_count", len(tenantIDs)))

	return nil
}

func (s *SimpleUserService) GetUser(ctx context.Context, userID domain.ID) (domain.User, error) {
	user, err := s.repository.GetByID(ctx, userID)
	if errors.Is(err, ErrUserNotFound) {
		slog.Warn("user not found", slog.String("user_id", userID.String()))
		return domain.User{}, ErrUserNotFound
	}
	if err != nil {
		slog.Error("getting user", slog.String("error", err.Error()))
		return domain.User{}, fmt.Errorf("getting user: %w", err)
	}

	return user, nil
}

func (s *SimpleUserService) FindByTenant(ctx context.Context, tenantID domain.ID) ([]domain.User, error) {
	users, err := s.repository.FindByTenant(ctx, tenantID)
	if err != nil {
		slog.Error("finding users by tenant", slog.String("error", err.Error()))
		return nil, fmt.Errorf("finding users by tenant: %w", err)
	}

	return users, nil
}
