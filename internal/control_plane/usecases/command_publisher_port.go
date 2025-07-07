package usecases

import (
	"context"
	"zensor-server/internal/shared_kernel/domain"
)

type CommandPublisher interface {
	Dispatch(context.Context, domain.Command) error
}
