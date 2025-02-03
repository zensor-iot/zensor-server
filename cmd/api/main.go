package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"zensor-server/internal/control_plane/httpapi"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/kafka"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/sages"

	"github.com/spf13/viper"
)

const (
	kafka_topic_device_registered string = "device_registered"
	kafka_topic_event_emitted     string = "event_emitted"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug})))
	slog.Info("🚀 zensor is initializing")

	config := loadConfig()
	slog.Debug("config loaded", "data", config)

	kafkaPublisherFactory := kafka.NewKafkaPublisherFactory(config.kafka.brokers)
	kafkaPublisherDevices, err := kafka.NewKafkaPublisher(config.kafka.brokers, config.kafka.topics["devices"], "")
	if err != nil {
		panic(err)
	}
	kafkaPublisherEventEmitted, err := kafka.NewKafkaPublisher(config.kafka.brokers, kafka_topic_event_emitted, "")
	if err != nil {
		panic(err)
	}

	c := make(chan mqtt.Event)
	go mqtt.Run(config.mqtt.broker, "dummy", c)
	go sages.EvaluateThingToServer(
		c,
		kafkaPublisherDevices,
		kafkaPublisherEventEmitted,
	)

	orm := initDatabase(config)
	deviceRepository, err := persistence.NewDeviceRepository(kafkaPublisherFactory, orm)
	if err != nil {
		panic(err)
	}
	deviceService := usecases.NewDeviceService(deviceRepository)
	deviceController := httpapi.NewDeviceController(deviceService)
	httpServer := httpserver.NewServer(
		deviceController,
	)
	go httpServer.Run()

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel
	slog.Info("good bye!!!")
	os.Exit(0)
}

func initDatabase(config appConfig) sql.ORM {
	db := sql.NewPosgreDatabase(config.postgresql.url)
	if err := db.Open(); err != nil {
		panic(err)
	}
	defer db.Close()
	dbMigration(db)

	orm, err := sql.NewPosgreORM(config.postgresql.dsn)
	if err != nil {
		panic(err)
	}

	return orm
}

func dbMigration(db sql.Database) {
	db.Up("migrations")

}

func loadConfig() appConfig {
	viper.SetEnvPrefix("zensor-server")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName("server")
	viper.AddConfigPath("config")
	viper.AddConfigPath("/config")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	return appConfig{
		mqtt: mqttConfig{
			broker: viper.GetString("mqtt.broker"),
		},
		postgresql: postgresqlConfig{
			url: viper.GetString("database.url"),
			dsn: viper.GetString("database.dsn"),
		},
		kafka: kafkaConfig{
			brokers: viper.GetStringSlice("kafka.brokers"),
			topics:  viper.GetStringMapString("kafka.topics"),
		},
	}
}

type appConfig struct {
	mqtt       mqttConfig
	kafka      kafkaConfig
	postgresql postgresqlConfig
}

type mqttConfig struct {
	broker string
}

type kafkaConfig struct {
	brokers []string
	topics  map[string]string
}

type postgresqlConfig struct {
	url string
	dsn string
}
