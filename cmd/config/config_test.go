package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfigWithRedis(t *testing.T) {
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
	if err := os.MkdirAll("config", 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Write temporary config file
	err := os.WriteFile("config/server_test.yaml", []byte(tempConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
	defer os.Remove("config/server_test.yaml")

	// Temporarily change config name
	originalConfigName := "server"
	defer func() {
		// Reset config instance and config name
		loadConfigOnce.Do(func() {}) // Reset the sync.Once
		viper.SetConfigName(originalConfigName)
	}()

	// Set config name to test file
	viper.SetConfigName("server_test")

	// Load config
	config := LoadConfig()

	// Verify Redis configuration
	if config.Redis.Addr != "localhost:6379" {
		t.Errorf("Expected Redis addr to be 'localhost:6379', got '%s'", config.Redis.Addr)
	}

	if config.Redis.Password != "" {
		t.Errorf("Expected Redis password to be empty, got '%s'", config.Redis.Password)
	}

	if config.Redis.DB != 0 {
		t.Errorf("Expected Redis DB to be 0, got %d", config.Redis.DB)
	}
}
