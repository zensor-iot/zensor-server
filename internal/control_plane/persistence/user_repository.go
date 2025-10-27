package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
	"zensor-server/internal/shared_kernel/domain"
)

const (
	_usersTopic = "users"
)

func NewUserRepository(publisherFactory pubsub.PublisherFactory, orm sql.ORM) (*SimpleUserRepository, error) {
	publisher, err := publisherFactory.New(_usersTopic, &avro.AvroUser{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	err = orm.AutoMigrate(&internal.User{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleUserRepository{
		publisher: publisher,
		orm:       orm,
	}, nil
}

var _ usecases.UserRepository = (*SimpleUserRepository)(nil)

type SimpleUserRepository struct {
	publisher pubsub.Publisher
	orm       sql.ORM
}

func (r *SimpleUserRepository) Upsert(ctx context.Context, user domain.User) error {
	tenantIDStrs := make([]string, len(user.Tenants))
	for i, id := range user.Tenants {
		tenantIDStrs[i] = id.String()
	}

	avroUser := &avro.AvroUser{
		ID:        user.ID.String(),
		Tenants:   tenantIDStrs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := r.publisher.Publish(ctx, pubsub.Key(user.ID), avroUser)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (r *SimpleUserRepository) GetByID(ctx context.Context, id domain.ID) (domain.User, error) {
	var entity internal.User
	err := r.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id.String()).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.User{}, usecases.ErrUserNotFound
	}

	if err != nil {
		return domain.User{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}
