package usecases

import (
	"context"
	"zensor-server/internal/shared_kernel/domain"
)

//go:generate mockgen -source=command_publisher_port.go -destination=../../../test/unit/doubles/control_plane/usecases/command_publisher_port_mock.go -package=usecases -mock_names=CommandPublisher=MockCommandPublisher

type CommandPublisher interface {
	Dispatch(context.Context, domain.Command) error
}
