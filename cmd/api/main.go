package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"zensor-server/internal/kafka"
	"zensor-server/internal/logger"
	"zensor-server/internal/mqtt"
	"zensor-server/internal/persistence"
	"zensor-server/internal/rest"
	"zensor-server/internal/sages"
	"zensor-server/internal/temperature"

	"github.com/spf13/viper"
)

const (
	kafka_topic_device_registered string = "device_registered"
	kafka_topic_event_emitted     string = "event_emitted"
)

var log *logger.Logger

func main() {
	logger.Info("🚀 zensor is initializing")

	config := loadConfig()
	logger.Info("config loaded", "data", config)

	db := persistence.NewDatabase(config.postgresql.url)
	db.Open()
	defer db.Close()
	dbMigration(db)
	eventRepository := persistence.NewEventRepository(db)

	restServer := rest.NewRestServer(eventRepository)

	consumer := kafka.NewKafkaConsumer(config.kafka.brokers, kafka_topic_device_registered)
	go logConsumedMessages(consumer)
	go temperatureFlow(consumer)

	kafkaPublisherDeviceRegistered, _ := kafka.NewKafkaPublisher(config.kafka.brokers, kafka_topic_device_registered)
	kafkaPublisherEventEmitted, _ := kafka.NewKafkaPublisher(config.kafka.brokers, kafka_topic_event_emitted)

	c := make(chan mqtt.Event)
	go mqtt.Run(config.mqtt.broker, config.mqtt.topic, c)
	go sages.EvaluateThingToServer(
		c,
		kafkaPublisherDeviceRegistered,
		kafkaPublisherEventEmitted,
	)
	go restServer.Run()

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel
	logger.Info("good bye!!!")
	os.Exit(0)
}

func logConsumedMessages(c kafka.KafkaConsumer) {
	c.Consume(func(msg string) {
		logger.Info(msg)
	})
}

func temperatureFlow(c kafka.KafkaConsumer) {
	handler := temperature.CreateHandler()
	c.Consume(func(msg string) {
		handler.Push(msg)
	})
}

func dbMigration(db persistence.Database) {
	db.Up("migrations")

}

func loadConfig() appConfig {
	viper.SetEnvPrefix("zensor-server")
	viper.AutomaticEnv()
	viper.SetConfigName("server")
	viper.AddConfigPath("config")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	return appConfig{
		mqtt: mqttConfig{
			broker: viper.GetString("mqtt.broker"),
			topic:  viper.GetString("mqtt.topic"),
		},
		postgresql: postgresqlConfig{
			url: viper.GetString("database.url"),
		},
		kafka: kafkaConfig{
			brokers: viper.GetStringSlice("kafka.brokers"),
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
	topic  string
}

type kafkaConfig struct {
	brokers []string
}

type postgresqlConfig struct {
	url string
}
