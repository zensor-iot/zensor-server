package communication

import (
	"context"
	"fmt"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/shared_kernel/avro"
)

const (
	_deviceCommandsTopic = "device_commands"
)

func NewCommandPublisher(factory pubsub.PublisherFactory) (*CommandPublisher, error) {
	publisher, err := factory.New(_deviceCommandsTopic, &avro.AvroCommand{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}
	return &CommandPublisher{
		publisher: publisher,
	}, nil
}

var _ usecases.CommandPublisher = (*CommandPublisher)(nil)

type CommandPublisher struct {
	publisher pubsub.Publisher
}

func (p *CommandPublisher) Dispatch(ctx context.Context, cmd domain.Command) error {
	avroCmd := avro.ToAvroCommand(cmd)
	avroCmd.Version++
	err := p.publisher.Publish(ctx, pubsub.Key(cmd.ID), avroCmd)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}
