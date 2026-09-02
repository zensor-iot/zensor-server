// Package persistence provides repository implementations for the medicines
// module.
package persistence

import (
	"context"
	"fmt"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/medicines/usecases"
)

// paginateByFilter is a copy of the maintenance module's helper. Promoting it
// to infra/sql would make that package depend on a usecases level Pagination
// type, which is a worse dependency than the duplication; a third occurrence
// would be the point to reconsider.
func paginateByFilter[T any](
	ctx context.Context,
	orm sql.ORM,
	model T,
	filter string,
	pagination usecases.Pagination,
	args ...any,
) ([]T, int, error) {
	var total int64
	query := orm.WithContext(ctx).Model(&model)

	if err := query.Where(filter, args...).Count(&total).Error(); err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []T
	if err := query.
		Where(filter, args...).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&entities).
		Error(); err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	return entities, int(total), nil
}
