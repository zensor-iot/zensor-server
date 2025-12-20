package config

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"zensor-server/internal/infra/utils"

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
				DSN:          viper.GetString("database.dsn"),
				QueryTimeout: viper.GetDuration("database.query_timeout"),
			},
			Kafka: KafkaConfig{
				Brokers:        viper.GetStringSlice("kafka.brokers"),
				Group:          viper.GetString("kafka.group"),
				SchemaRegistry: viper.GetString("kafka.schema_registry"),
			},
			Redis: RedisConfig{
				Addr:     viper.GetString("redis.addr"),
				Password: viper.GetString("redis.password"),
				DB:       viper.GetInt("redis.db"),
			},
			MailerSend: MailerSendConfig{
				APIKey:    viper.GetString("mailersend.api_key"),
				FromEmail: viper.GetString("mailersend.from_email"),
				FromName:  viper.GetString("mailersend.from_name"),
			},
			Metrics: loadMetricsConfig(),
			Modules: loadModulesConfig(),
		}
	})

	return configInstance
}

func loadMetricsConfig() []MetricWorkerConfig {
	metricsInterface := viper.Get("metrics")
	if metricsSlice, ok := metricsInterface.([]interface{}); ok {
		var metrics []MetricWorkerConfig
		for _, item := range metricsSlice {
			if metricMap, ok := item.(map[string]interface{}); ok {
				metric := MetricWorkerConfig{
					Name:              utils.ExtractStringValue(metricMap, "name"),
					Type:              utils.ExtractStringValue(metricMap, "type"),
					Topic:             utils.ExtractStringValue(metricMap, "topic"),
					EventType:         utils.ExtractStringValue(metricMap, "event_type"),
					ValuePropertyName: utils.ExtractStringValue(metricMap, "value_property_name"),
					CustomAttributes:  make(map[string]string),
				}
				if customAttributes, exists := metricMap["custom_attributes"]; exists {
					if attrMap, ok := customAttributes.(map[string]any); ok {
						for k, v := range attrMap {
							metric.CustomAttributes[k] = fmt.Sprintf("%v", v)
						}
					}
				}
				metrics = append(metrics, metric)
			}
		}
		return metrics
	}
	return []MetricWorkerConfig{}
}

func loadModulesConfig() ModulesConfig {
	return ModulesConfig{
		Permaculture: ModuleConfig{
			Enabled: viper.GetBool("modules.permaculture.enabled"),
		},
		Maintenance: ModuleConfig{
			Enabled: viper.GetBool("modules.maintenance.enabled"),
		},
	}
}

type AppConfig struct {
	General    GeneralConfig
	mqtt       MqttConfig
	MQTTClient MQTTClientConfig
	Kafka      KafkaConfig
	Postgresql PostgresqlConfig
	Redis      RedisConfig
	MailerSend MailerSendConfig
	Metrics    MetricsConfig
	Modules    ModulesConfig
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
	DSN          string
	QueryTimeout time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type MailerSendConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
}

type MetricsConfig []MetricWorkerConfig

type MetricWorkerConfig struct {
	Name              string
	Type              string
	Topic             string
	EventType         string
	ValuePropertyName string
	CustomAttributes  map[string]string
}

type ModulesConfig struct {
	Permaculture ModuleConfig
	Maintenance  ModuleConfig
}

type ModuleConfig struct {
	Enabled bool
}
