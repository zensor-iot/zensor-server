package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var loadConfigOnce sync.Once
var configInstance AppConfig

func LoadConfig() AppConfig {
	loadConfigOnce.Do(func() {
		viper.SetEnvPrefix("zensor_server")
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.SetConfigName("server")
		viper.AddConfigPath("config")
		viper.AddConfigPath("/config")
		if err := viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
		configInstance = AppConfig{
			General: GeneralConfig{
				LogLevel: viper.GetString("general.log_level"),
			},
			mqtt: MqttConfig{
				Broker: viper.GetString("mqtt.broker"),
			},
			MQTTClient: MQTTClientConfig{
				Broker:   viper.GetString("mqtt_client.broker"),
				ClientID: viper.GetString("mqtt_client.client_id"),
				Username: viper.GetString("mqtt_client.username"),
				Password: viper.GetString("mqtt_client.password"),
			},
			Postgresql: PostgresqlConfig{
				URL:                   viper.GetString("database.url"),
				DSN:                   viper.GetString("database.dsn"),
				MigrationReplacements: viper.GetStringMapString("database.migration_replacements"),
			},
			Kafka: KafkaConfig{
				Brokers:        viper.GetStringSlice("kafka.brokers"),
				Group:          viper.GetString("kafka.group"),
				SchemaRegistry: viper.GetString("kafka.schema_registry"),
			},
		}
	})

	return configInstance
}

type AppConfig struct {
	General    GeneralConfig
	mqtt       MqttConfig
	MQTTClient MQTTClientConfig
	Kafka      KafkaConfig
	Postgresql PostgresqlConfig
}

type GeneralConfig struct {
	LogLevel string
}

type MqttConfig struct {
	Broker string
}

type MQTTClientConfig struct {
	Broker   string
	ClientID string
	Username string
	Password string
}

type KafkaConfig struct {
	Brokers        []string
	Group          string
	SchemaRegistry string
}

type PostgresqlConfig struct {
	URL                   string
	DSN                   string
	MigrationReplacements map[string]string
}
