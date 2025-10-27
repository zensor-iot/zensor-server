package wire

import (
	"zensor-server/cmd/config"
	"zensor-server/internal/infra/replication/handlers"
)

func provideAppConfig() config.AppConfig {
	return config.LoadConfig()
}

func InitializeTenantConfigurationHandler() (*handlers.TenantConfigurationHandler, error) {
	return handlers.NewTenantConfigurationHandler(provideDatabase(provideAppConfig())), nil
}

func InitializeUserHandler() (*handlers.UserHandler, error) {
	return handlers.NewUserHandler(provideDatabase(provideAppConfig())), nil
}
