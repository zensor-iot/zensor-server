package usecases

import (
	"context"
	"errors"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

var (
	ErrActivityNotFound  = errors.New("activity not found")
	ErrExecutionNotFound = errors.New("execution not found")
)

type Pagination struct {
	Limit  int
	Offset int
}

type ActivityRepository interface {
	Create(ctx context.Context, activity maintenanceDomain.Activity) error
	GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Activity, error)
	FindAllByTenant(ctx context.Context, tenantID shareddomain.ID, pagination Pagination) ([]maintenanceDomain.Activity, int, error)
	Update(ctx context.Context, activity maintenanceDomain.Activity) error
	Delete(ctx context.Context, id shareddomain.ID) error
}

type ExecutionRepository interface {
	Create(ctx context.Context, execution maintenanceDomain.Execution) error
	GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Execution, error)
	FindAllByActivity(ctx context.Context, activityID shareddomain.ID, pagination Pagination) ([]maintenanceDomain.Execution, int, error)
	Update(ctx context.Context, execution maintenanceDomain.Execution) error
	MarkCompleted(ctx context.Context, id shareddomain.ID, completedBy string) error
	FindAllOverdue(ctx context.Context, tenantID shareddomain.ID) ([]maintenanceDomain.Execution, error)
	FindAllDueSoon(ctx context.Context, tenantID shareddomain.ID, days int) ([]maintenanceDomain.Execution, error)
}
