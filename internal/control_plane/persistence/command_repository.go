package persistence

import (
	"context"
	"fmt"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
	"zensor-server/internal/shared_kernel/domain"
)

func NewCommandRepository(
	orm sql.ORM,
	publisherFactory pubsub.PublisherFactory,
) (*SimpleCommandRepository, error) {
	err := orm.AutoMigrate(&internal.Command{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating command: %w", err)
	}

	commandPublisher, err := publisherFactory.New("device_commands", &avro.AvroCommand{})
	if err != nil {
		return nil, fmt.Errorf("creating command publisher: %w", err)
	}

	return &SimpleCommandRepository{
		orm:              orm,
		commandPublisher: commandPublisher,
	}, nil
}

var _ usecases.CommandRepository = (*SimpleCommandRepository)(nil)

type SimpleCommandRepository struct {
	orm              sql.ORM
	commandPublisher pubsub.Publisher
}

func (r *SimpleCommandRepository) Create(ctx context.Context, cmd domain.Command) error {
	avroCmd := avro.ToAvroCommand(cmd)
	err := r.commandPublisher.Publish(ctx, pubsub.Key(cmd.ID), avroCmd)
	if err != nil {
		return fmt.Errorf("publishing command to kafka: %w", err)
	}

	return nil
}

func (r *SimpleCommandRepository) Update(ctx context.Context, cmd domain.Command) error {
	var existingCmd internal.Command
	err := r.orm.WithContext(ctx).First(&existingCmd, "id = ?", cmd.ID.String()).Error()
	if err != nil {
		return fmt.Errorf("command not found in database: %w", err)
	}

	cmd.CreatedAt = existingCmd.CreatedAt

	if cmd.Status == "" {
		cmd.Status = domain.CommandStatus(existingCmd.Status)
	}

	if cmd.QueuedAt == nil && existingCmd.QueuedAt != nil {
		cmd.QueuedAt = existingCmd.QueuedAt
	}
	if cmd.AckedAt == nil && existingCmd.AckedAt != nil {
		cmd.AckedAt = existingCmd.AckedAt
	}
	if cmd.FailedAt == nil && existingCmd.FailedAt != nil {
		cmd.FailedAt = existingCmd.FailedAt
	}

	avroCmd := avro.ToAvroCommand(cmd)
	avroCmd.Version++
	err = r.commandPublisher.Publish(ctx, pubsub.Key(cmd.ID), avroCmd)
	if err != nil {
		return fmt.Errorf("publishing command update to kafka: %w", err)
	}

	return nil
}

func (r *SimpleCommandRepository) FindAllPending(ctx context.Context) ([]domain.Command, error) {
	var entities internal.CommandSet
	err := r.orm.
		WithContext(ctx).
		Where("sent = ? AND ready = ?", false, false).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	return entities.ToDomain(), nil
}

func (r *SimpleCommandRepository) FindPendingByDevice(ctx context.Context, deviceID domain.ID) ([]domain.Command, error) {
	var entities internal.CommandSet
	err := r.orm.
		WithContext(ctx).
		Where("sent = ? AND device_id = ?", false, deviceID.String()).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	return entities.ToDomain(), nil
}

func (r *SimpleCommandRepository) GetByID(ctx context.Context, id domain.ID) (domain.Command, error) {
	var entity internal.Command
	err := r.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id.String()).
		Error()

	if err != nil {
		return domain.Command{}, fmt.Errorf("database query: %w", err)
	}

	domainCmd := entity.ToDomain()
	return domainCmd, nil
}

func (r *SimpleCommandRepository) FindByTaskID(ctx context.Context, taskID domain.ID) ([]domain.Command, error) {
	var entities internal.CommandSet
	err := r.orm.
		WithContext(ctx).
		Where("task_id = ?", taskID.String()).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	return entities.ToDomain(), nil
}
