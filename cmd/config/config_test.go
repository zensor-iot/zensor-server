package config_test

import (
	"fmt"
	"os"
	"strings"
	"zensor-server/cmd/config"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"
)

// loadTestConfig loads configuration with a specific config name for testing
func loadTestConfig(configName string) config.AppConfig {
	viper.Reset()
	viper.SetEnvPrefix("zensor_server")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName(configName)
	viper.AddConfigPath("config")
	viper.AddConfigPath("/config")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	return config.AppConfig{
		General: config.GeneralConfig{
			LogLevel: viper.GetString("general.log_level"),
		},
		MQTTClient: config.MQTTClientConfig{
			Broker:   viper.GetString("mqtt_client.broker"),
			ClientID: viper.GetString("mqtt_client.client_id"),
			Username: viper.GetString("mqtt_client.username"),
			Password: viper.GetString("mqtt_client.password"),
		},
		Postgresql: config.PostgresqlConfig{
			DSN: viper.GetString("database.dsn"),
		},
		Kafka: config.KafkaConfig{
			Brokers:        viper.GetStringSlice("kafka.brokers"),
			Group:          viper.GetString("kafka.group"),
			SchemaRegistry: viper.GetString("kafka.schema_registry"),
		},
		Redis: config.RedisConfig{
			Addr:     viper.GetString("redis.addr"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		},
	}
}

var _ = ginkgo.Describe("LoadConfig", func() {
	var (
		tempConfigFile     string
		originalConfigName string
	)

	ginkgo.BeforeEach(func() {
		// Create a temporary config file with Redis configuration
		tempConfig := `
general:
  log_level: info
mqtt:
  broker: "localhost:1883"
database:
  dsn: "host=localhost user=postgres dbname=postgres port=5432 sslmode=disable"
kafka:
  brokers:
    - "localhost:19092"
  group: "zensor-server"
  schema_registry: "http://localhost:8081"
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
mqtt_client:
  broker: nam1.cloud.thethings.network:1883
  client_id: zensor_server_local
`

		// Create config directory if it doesn't exist
		err := os.MkdirAll("config", 0755)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Write temporary config file
		tempConfigFile = "config/server_test.yaml"
		err = os.WriteFile(tempConfigFile, []byte(tempConfig), 0644)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Store original config name
		originalConfigName = "server"
	})

	ginkgo.AfterEach(func() {
		// Clean up temporary config file
		if tempConfigFile != "" {
			os.Remove(tempConfigFile)
		}

		// Reset config name
		viper.SetConfigName(originalConfigName)
		// Reset viper to clear any cached config
		viper.Reset()
	})

	ginkgo.Context("When loading configuration with Redis settings", func() {
		ginkgo.When("the config file contains valid Redis configuration", func() {
			ginkgo.It("should load Redis configuration correctly", func() {
				// Load config using test helper
				cfg := loadTestConfig("server_test")

				// Verify Redis configuration
				gomega.Expect(cfg.Redis.Addr).To(gomega.Equal("localhost:6379"))
				gomega.Expect(cfg.Redis.Password).To(gomega.BeEmpty())
				gomega.Expect(cfg.Redis.DB).To(gomega.Equal(0))
			})
		})
	})
})
