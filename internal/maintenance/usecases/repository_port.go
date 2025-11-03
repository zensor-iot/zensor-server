package usecases

import (
	"context"
	"errors"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

var (
	ErrMaintenanceActivityNotFound  = errors.New("maintenance activity not found")
	ErrMaintenanceExecutionNotFound = errors.New("maintenance execution not found")
)

type Pagination struct {
	Limit  int
	Offset int
}

type MaintenanceActivityRepository interface {
	Create(ctx context.Context, activity maintenanceDomain.MaintenanceActivity) error
	GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.MaintenanceActivity, error)
	FindAllByTenant(ctx context.Context, tenantID shareddomain.ID, pagination Pagination) ([]maintenanceDomain.MaintenanceActivity, int, error)
	Update(ctx context.Context, activity maintenanceDomain.MaintenanceActivity) error
	Delete(ctx context.Context, id shareddomain.ID) error
}

type MaintenanceExecutionRepository interface {
	Create(ctx context.Context, execution maintenanceDomain.MaintenanceExecution) error
	GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.MaintenanceExecution, error)
	FindAllByActivity(ctx context.Context, activityID shareddomain.ID, pagination Pagination) ([]maintenanceDomain.MaintenanceExecution, int, error)
	Update(ctx context.Context, execution maintenanceDomain.MaintenanceExecution) error
	MarkCompleted(ctx context.Context, id shareddomain.ID, completedBy string) error
	FindAllOverdue(ctx context.Context, tenantID shareddomain.ID) ([]maintenanceDomain.MaintenanceExecution, error)
	FindAllDueSoon(ctx context.Context, tenantID shareddomain.ID, days int) ([]maintenanceDomain.MaintenanceExecution, error)
}
