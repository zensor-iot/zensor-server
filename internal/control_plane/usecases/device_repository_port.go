package usecases

import (
	"context"
	"zensor-server/internal/control_plane/domain"
)

type CommandPublisher interface {
	Dispatch(context.Context, domain.Command) error
}
