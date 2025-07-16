package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
	"zensor-server/internal/shared_kernel/domain"
)

func NewTenantConfigurationRepository(publisherFactory pubsub.PublisherFactory, orm sql.ORM) (*SimpleTenantConfigurationRepository, error) {
	publisher, err := publisherFactory.New("tenant_configurations", &avro.AvroTenantConfiguration{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	err = orm.AutoMigrate(&internal.TenantConfiguration{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleTenantConfigurationRepository{
		publisher: publisher,
		orm:       orm,
	}, nil
}

var _ usecases.TenantConfigurationRepository = (*SimpleTenantConfigurationRepository)(nil)

type SimpleTenantConfigurationRepository struct {
	publisher pubsub.Publisher
	orm       sql.ORM
}

func (r *SimpleTenantConfigurationRepository) Create(ctx context.Context, config domain.TenantConfiguration) error {
	// Convert domain config to Avro config
	avroConfig := &avro.AvroTenantConfiguration{
		ID:        config.ID.String(),
		TenantID:  config.TenantID.String(),
		Timezone:  config.Timezone,
		Version:   config.Version,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}

	err := r.publisher.Publish(ctx, pubsub.Key(config.ID), avroConfig)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (r *SimpleTenantConfigurationRepository) GetByTenantID(ctx context.Context, tenantID domain.ID) (domain.TenantConfiguration, error) {
	var entity internal.TenantConfiguration
	err := r.orm.
		WithContext(ctx).
		Where("tenant_id = ?", tenantID.String()).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.TenantConfiguration{}, usecases.ErrTenantConfigurationNotFound
	}

	if err != nil {
		return domain.TenantConfiguration{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleTenantConfigurationRepository) Update(ctx context.Context, config domain.TenantConfiguration) error {
	// Convert domain config to Avro config
	avroConfig := &avro.AvroTenantConfiguration{
		ID:        config.ID.String(),
		TenantID:  config.TenantID.String(),
		Timezone:  config.Timezone,
		Version:   config.Version,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}

	err := r.publisher.Publish(ctx, pubsub.Key(config.ID), avroConfig)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}
