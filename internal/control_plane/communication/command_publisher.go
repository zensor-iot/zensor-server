package communication

import (
	"context"
	"fmt"
	"zensor-server/internal/control_plane/communication/internal"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
)

const (
	_deviceCommandsTopic = "device_commands"
)

func NewCommandPublisher(factory pubsub.PublisherFactory) (*CommandPublisher, error) {
	publisher, err := factory.New(_deviceCommandsTopic, internal.Command{})
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
	internalCmd := internal.FromCommand(cmd)
	err := p.publisher.Publish(ctx, pubsub.Key(cmd.ID), internalCmd)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}
