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

	"zensor-server/internal/control_plane/communication"
	"zensor-server/internal/control_plane/httpapi"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/workers"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"

	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
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

	shutdownOtel := startOTel()

	kafkaPublisherFactory := pubsub.NewKafkaPublisherFactory(config.kafka.brokers)
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

	commandPublisher, err := communication.NewCommandPublisher(kafkaPublisherFactory)
	if err != nil {
		panic(err)
	}

	deviceService := usecases.NewDeviceService(deviceRepository, commandPublisher)
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
	internalBroker := async.NewLocalBroker()
	consumerFactory := pubsub.NewKafkaConsumerFactory(config.kafka.brokers, config.kafka.group)
	wg.Add(1)
	workers.NewLoraIntegrationWorker(ticker, deviceService, mqttClient, internalBroker, consumerFactory).Run(appCtx, wg.Done)

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel
	shutdownOtel()
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
			group:   viper.GetString("kafka.group"),
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
	group   string
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

type ShutdownFunc func() error

const (
	_defautlEndpoint = "localhost:4317"
	_collectPeriod   = 30 * time.Second
	_collectTimeout  = 35 * time.Second
	_minimumInterval = time.Minute
)

var (
	_histogramBuckets = []float64{5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000, 25000, 50000, 100000}
)

func startOTel() ShutdownFunc {
	slog.Info("starting OTel providers")
	shutdown, err := otelStart(context.Background())
	if err != nil {
		panic(err)
	}

	return shutdown
}

func otelStart(ctx context.Context) (ShutdownFunc, error) {
	metricsShutdownFunc, err := startMetricsProvider(ctx)
	if err != nil {
		return nil, err
	}

	return func() error {
		if err := metricsShutdownFunc(); err != nil {
			return err
		}

		return nil
	}, nil
}

func startMetricsProvider(ctx context.Context) (ShutdownFunc, error) {
	exp, err := newMetricExporter(ctx)
	if err != nil {
		return nil, err
	}

	mp := newMeterProvider(exp)
	otel.SetMeterProvider(mp)

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(_minimumInterval))
	if err != nil {
		return nil, err
	}

	return func() error {
		return mp.Shutdown(ctx)
	}, nil
}

func newMetricExporter(ctx context.Context) (metric.Exporter, error) {
	return otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(_defautlEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
}

func newMeterProvider(metricExporter metric.Exporter) *metric.MeterProvider {
	return metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(
				metricExporter,
				metric.WithTimeout(_collectTimeout),
				metric.WithInterval(_collectPeriod))),
		metric.WithView(metric.NewView(
			metric.Instrument{
				Name: "*",
				Kind: metric.InstrumentKindHistogram,
			},
			metric.Stream{
				Aggregation: metric.AggregationExplicitBucketHistogram{
					Boundaries: _histogramBuckets,
				},
			},
		)),
	)
}
