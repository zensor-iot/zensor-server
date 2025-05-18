package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"zensor-server/cmd/api/wire"
	"zensor-server/cmd/config"
	"zensor-server/internal/data_plane/workers"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/pubsub"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug, ReplaceAttr: slogReplaceSource})))
	slog.Info("🚀 zensor is initializing")

	config := config.LoadConfig()
	slog.Debug("config loaded", "data", config)

	shutdownOtel := startOTel()

	httpServer := httpserver.NewServer(
		handleWireInjector(wire.InitializeDeviceController()).(httpserver.Controller),
		handleWireInjector(wire.InitializeEvaluationRuleController()).(httpserver.Controller),
	)

	appCtx, cancelFn := context.WithCancel(context.Background())
	go httpServer.Run()

	// DATA PLANE
	var wg sync.WaitGroup
	ticker := time.NewTicker(30 * time.Second)
	simpleClientOpts := mqtt.SimpleClientOpts{
		Broker:   config.MQTTClient.Broker,
		ClientID: config.MQTTClient.ClientID,
		Username: config.MQTTClient.Username,
		Password: config.MQTTClient.Password, //pragma: allowlist secret
	}
	mqttClient := mqtt.NewSimpleClient(simpleClientOpts)
	internalBroker := async.NewLocalBroker()
	consumerFactory := pubsub.NewKafkaConsumerFactory(config.Kafka.Brokers, config.Kafka.Group)
	deviceService, err := wire.InitializeDeviceService()
	if err != nil {
		panic(err)
	}

	// TODO: capture workers into a variable to shutdown them later
	wg.Add(1)
	go workers.NewLoraIntegrationWorker(ticker, deviceService, mqttClient, internalBroker, consumerFactory).Run(appCtx, wg.Done)
	wg.Add(1)
	go handleWireInjector(wire.InitializeCommandWorker(internalBroker)).(async.Worker).Run(appCtx, wg.Done)

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel
	shutdownOtel()
	cancelFn()
	wg.Wait()
	slog.Info("good bye!!!")
	os.Exit(0)
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
	endpoint := _defautlEndpoint
	if value, ok := os.LookupEnv("ZENSOR_SERVER_OTELCOL_ENDPOINT"); ok {
		endpoint = value
	}

	return otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
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

func handleWireInjector(value any, err error) any {
	if err != nil {
		panic(err)
	}

	return value
}
