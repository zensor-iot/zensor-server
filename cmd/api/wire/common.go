package wire

import (
	"zensor-server/cmd/config"
)

func provideAppConfig() config.AppConfig {
	return config.LoadConfig()
}
