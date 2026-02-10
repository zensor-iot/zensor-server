package persistence

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"zensor-server/internal/shared_kernel/persistence/internal"
	"zensor-server/internal/shared_kernel/usecases"
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

	slog.Info("publishing user to pubsub", slog.String("user_id", user.ID.String()), slog.Any("tenants", tenantIDStrs))
	err := r.publisher.Publish(ctx, pubsub.Key(user.ID), avroUser)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}
	slog.Info("published user to pubsub", slog.String("user_id", user.ID.String()))

	return nil
}

func (r *SimpleUserRepository) GetByID(ctx context.Context, id domain.ID) (domain.User, error) {
	slog.Info("getting user by ID", slog.String("user_id", id.String()))
	var entity internal.User
	err := r.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id.String()).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		slog.Warn("user not found in database", slog.String("user_id", id.String()))
		return domain.User{}, usecases.ErrUserNotFound
	}

	if err != nil {
		slog.Error("database query error", slog.String("error", err.Error()))
		return domain.User{}, fmt.Errorf("database query: %w", err)
	}

	slog.Info("found user in database", slog.String("user_id", id.String()), slog.Any("tenants", entity.Tenants))
	return entity.ToDomain(), nil
}

func (r *SimpleUserRepository) FindByTenant(ctx context.Context, tenantID domain.ID) ([]domain.User, error) {
	var entities []internal.User
	err := r.orm.
		WithContext(ctx).
		Where("? = ANY(tenants)", tenantID.String()).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	users := make([]domain.User, len(entities))
	for i, entity := range entities {
		users[i] = entity.ToDomain()
	}

	return users, nil
}
