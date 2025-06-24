package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
)

func NewTenantRepository(publisherFactory pubsub.PublisherFactory, orm sql.ORM) (*SimpleTenantRepository, error) {
	publisher, err := publisherFactory.New("tenants", internal.Tenant{})
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
	data := internal.FromTenant(tenant)
	err := r.publisher.Publish(ctx, pubsub.Key(tenant.ID), data)
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
	data := internal.FromTenant(tenant)

	err := r.publisher.Publish(ctx, pubsub.Key(tenant.ID), data)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (r *SimpleTenantRepository) FindAll(ctx context.Context, includeDeleted bool) ([]domain.Tenant, error) {
	var entities []internal.Tenant

	query := r.orm.WithContext(ctx)

	// Filter out soft-deleted tenants unless specifically requested
	if !includeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	err := query.Find(&entities).Error()
	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	result := make([]domain.Tenant, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, nil
}
