package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"zensor-server/internal/control_plane/httpapi"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/workers"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/kafka"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/sql"

	"github.com/spf13/viper"
)

const (
	kafka_topic_device_registered string = "device_registered"
	kafka_topic_event_emitted     string = "event_emitted"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug, ReplaceAttr: slogReplaceSource})))
	slog.Info("🚀 zensor is initializing")

	config := loadConfig()
	slog.Debug("config loaded", "data", config)

	kafkaPublisherFactory := kafka.NewKafkaPublisherFactory(config.kafka.brokers)
	// kafkaPublisherDevices, err := kafka.NewKafkaPublisher(config.kafka.brokers, config.kafka.topics["devices"], "")
	// if err != nil {
	// 	panic(err)
	// }
	// kafkaPublisherEventEmitted, err := kafka.NewKafkaPublisher(config.kafka.brokers, kafka_topic_event_emitted, "")
	// if err != nil {
	// 	panic(err)
	// }

	// c := make(chan mqtt.Event)
	//go mqtt.Run(config.mqtt.broker, "dummy", c)
	// go sages.EvaluateThingToServer(
	// 	c,
	// 	kafkaPublisherDevices,
	// 	kafkaPublisherEventEmitted,
	// )

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

	appCtx, cancelFn := context.WithCancel(context.Background())
	go httpServer.Run()

	// DATA PLANE
	var wg sync.WaitGroup
	ticker := time.NewTicker(30 * time.Second)
	simpleClientOpts := mqtt.SimpleClientOpts{
		Broker:   config.mqttClient.broker,
		ClientID: config.mqttClient.clientID,
		Username: config.mqttClient.username,
		Password: config.mqttClient.password, //pragma: allowlist secret
	}
	mqttClient := mqtt.NewSimpleClient(simpleClientOpts)
	wg.Add(1)
	workers.NewLoraIntegrationWorker(ticker, deviceService, mqttClient).Run(appCtx, wg.Done)

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel
	cancelFn()
	wg.Wait()
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
	viper.SetEnvPrefix("zensor_server")
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
		mqttClient: mqttClientConfig{
			broker:   viper.GetString("mqtt_client.broker"),
			clientID: viper.GetString("mqtt_client.client_id"),
			username: viper.GetString("mqtt_client.username"),
			password: viper.GetString("mqtt_client.password"),
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
	mqttClient mqttClientConfig
	kafka      kafkaConfig
	postgresql postgresqlConfig
}

type mqttConfig struct {
	broker string
}

type mqttClientConfig struct {
	broker   string
	clientID string
	username string
	password string
}

type kafkaConfig struct {
	brokers []string
	topics  map[string]string
}

type postgresqlConfig struct {
	url string
	dsn string
}

func slogReplaceSource(groups []string, a slog.Attr) slog.Attr {
	// Check if the attribute is the source key
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		// Set the file attribute to only its base name
		source.File = filepath.Base(source.File)
		return slog.Any(a.Key, source)
	}
	return a // Return unchanged attribute for others
}
