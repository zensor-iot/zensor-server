package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
)

func NewTenantRepository(publisherFactory pubsub.PublisherFactory, orm sql.ORM) (*SimpleTenantRepository, error) {
	publisher, err := publisherFactory.New("tenants", &avro.AvroTenant{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	err = orm.AutoMigrate(&internal.Tenant{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleTenantRepository{
		publisher: publisher,
		orm:       orm,
	}, nil
}

var _ usecases.TenantRepository = (*SimpleTenantRepository)(nil)

type SimpleTenantRepository struct {
	publisher pubsub.Publisher
	orm       sql.ORM
}

func (r *SimpleTenantRepository) Create(ctx context.Context, tenant domain.Tenant) error {
	// Convert domain tenant to Avro tenant
	avroTenant := &avro.AvroTenant{
		ID:          tenant.ID.String(),
		Version:     tenant.Version,
		Name:        tenant.Name,
		Email:       tenant.Email,
		Description: tenant.Description,
		IsActive:    tenant.IsActive,
		CreatedAt:   tenant.CreatedAt,
		UpdatedAt:   tenant.UpdatedAt,
	}

	if tenant.DeletedAt != nil {
		avroTenant.DeletedAt = tenant.DeletedAt
	}

	err := r.publisher.Publish(ctx, pubsub.Key(tenant.ID), avroTenant)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (r *SimpleTenantRepository) GetByID(ctx context.Context, id domain.ID) (domain.Tenant, error) {
	var entity internal.Tenant
	err := r.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id.String()).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.Tenant{}, usecases.ErrTenantNotFound
	}

	if err != nil {
		return domain.Tenant{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleTenantRepository) GetByName(ctx context.Context, name string) (domain.Tenant, error) {
	var entity internal.Tenant
	err := r.orm.
		WithContext(ctx).
		Where("name = ?", name).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.Tenant{}, usecases.ErrTenantNotFound
	}

	if err != nil {
		return domain.Tenant{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleTenantRepository) Update(ctx context.Context, tenant domain.Tenant) error {
	// Convert domain tenant to Avro tenant
	avroTenant := &avro.AvroTenant{
		ID:          tenant.ID.String(),
		Version:     tenant.Version,
		Name:        tenant.Name,
		Email:       tenant.Email,
		Description: tenant.Description,
		IsActive:    tenant.IsActive,
		CreatedAt:   tenant.CreatedAt,
		UpdatedAt:   tenant.UpdatedAt,
	}

	if tenant.DeletedAt != nil {
		avroTenant.DeletedAt = tenant.DeletedAt
	}

	err := r.publisher.Publish(ctx, pubsub.Key(tenant.ID), avroTenant)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (r *SimpleTenantRepository) FindAll(ctx context.Context, includeDeleted bool, pagination usecases.Pagination) ([]domain.Tenant, int, error) {
	var total int64
	query := r.orm.WithContext(ctx).Model(&internal.Tenant{})

	// Filter out soft-deleted tenants unless specifically requested
	if !includeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	err := query.Count(&total).Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.Tenant
	query = r.orm.WithContext(ctx)

	// Filter out soft-deleted tenants unless specifically requested
	if !includeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	err = query.Limit(pagination.Limit).Offset(pagination.Offset).Find(&entities).Error()
	if err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	result := make([]domain.Tenant, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, int(total), nil
}
