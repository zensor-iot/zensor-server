//go:build wireinject
// +build wireinject

package wire

import (
	"zensor-server/internal/infra/replication/handlers"

	"github.com/google/wire"
)

func InitializeTenantConfigurationHandler() (*handlers.TenantConfigurationHandler, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		handlers.NewTenantConfigurationHandler,
	)
	return nil, nil
}

func InitializeUserHandler() (*handlers.UserHandler, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		handlers.NewUserHandler,
	)
	return nil, nil
}
