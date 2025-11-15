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
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/node"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/replication"
	"zensor-server/internal/infra/replication/handlers"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var (
	logLevelMapping = map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
)

func main() {
	config := config.LoadConfig()

	level := logLevelMapping[config.General.LogLevel]
	baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: level, ReplaceAttr: slogReplaceAttr})
	handler := baseHandler.WithAttrs([]slog.Attr{slog.String("version", node.Version)})
	slog.SetDefault(slog.New(handler))
	slog.Info("🚀 zensor is initializing")
	slog.Debug("config loaded", "data", config)

	shutdownOtel := startOTel()

	// DATA PLANE - Set up broker and dependencies first
	internalBroker := async.NewLocalBroker()

	httpServer := httpserver.NewServer(
		handleWireInjector(wire.InitializeDeviceController()).(httpserver.Controller),
		handleWireInjector(wire.InitializeEvaluationRuleController()).(httpserver.Controller),
		handleWireInjector(wire.InitializeTaskController()).(httpserver.Controller),
		handleWireInjector(wire.InitializeTenantController()).(httpserver.Controller),
		handleWireInjector(wire.InitializeTenantConfigurationController()).(httpserver.Controller),
		handleWireInjector(wire.InitializeScheduledTaskController()).(httpserver.Controller),
		handleWireInjector(wire.InitializeUserController()).(httpserver.Controller),
		handleWireInjector(wire.InitializeDeviceMessageWebSocketController(internalBroker)).(httpserver.Controller),
		handleWireInjector(wire.InitializeDeviceSpecificWebSocketController(internalBroker)).(httpserver.Controller),
		handleWireInjector(wire.InitializeMaintenanceActivityController()).(httpserver.Controller),
		handleWireInjector(wire.InitializeMaintenanceExecutionController()).(httpserver.Controller),
	)

	appCtx, cancelFn := context.WithCancel(context.Background())
	go httpServer.Run()

	// Initialize replication service for local environment
	replicationService := initializeReplicationService()

	var wg sync.WaitGroup
	ticker := time.NewTicker(30 * time.Second)
	simpleClientOpts := mqtt.SimpleClientOpts{
		Broker:   config.MQTTClient.Broker,
		ClientID: config.MQTTClient.ClientID,
		Username: config.MQTTClient.Username,
		Password: config.MQTTClient.Password, //pragma: allowlist secret
	}
	mqttClient := mqtt.NewSimpleClient(simpleClientOpts)

	// Use environment-aware consumer factory
	var consumerFactory pubsub.ConsumerFactory
	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "production"
	}
	if env == "local" {
		consumerFactory = pubsub.NewMemoryConsumerFactory("lora-integration")
	} else {
		consumerFactory = pubsub.NewKafkaConsumerFactory(config.Kafka.Brokers, config.Kafka.Group, config.Kafka.SchemaRegistry)
	}

	// TODO: capture workers into a variable to shutdown them later
	wg.Add(1)
	go handleWireInjector(wire.InitializeLoraIntegrationWorker(ticker, mqttClient, internalBroker, consumerFactory)).(async.Worker).Run(appCtx, wg.Done)
	wg.Add(1)
	go handleWireInjector(wire.InitializeCommandWorker(internalBroker)).(async.Worker).Run(appCtx, wg.Done)
	wg.Add(1)
	go handleWireInjector(wire.InitializeScheduledTaskWorker(internalBroker)).(async.Worker).Run(appCtx, wg.Done)
	wg.Add(1)
	go handleWireInjector(wire.InitializeNotificationWorker(internalBroker)).(async.Worker).Run(appCtx, wg.Done)

	// Initialize metric workers based on configuration
	metricWorkerFactory := wire.InitializeMetricWorkerFactory(internalBroker)
	metricWorkers, err := metricWorkerFactory.CreateWorkers(config.Metrics)
	if err != nil {
		slog.Error("failed to create metric workers", slog.Any("error", err))
		panic(err)
	}

	// Start all metric workers
	for _, worker := range metricWorkers {
		wg.Add(1)
		go worker.Run(appCtx, wg.Done)
	}

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel
	shutdownOtel()

	// Stop replication service if running
	if replicationService != nil {
		replicationService.Stop()
	}

	cancelFn()
	wg.Wait()
	slog.Info("good bye!!!")
	os.Exit(0)
}

// initializeReplicationService initializes and starts the replication service for local environment
func initializeReplicationService() *replication.Service {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "production"
	}

	if env != "local" {
		return nil
	}

	slog.Info("initializing replication service for local environment")

	// Initialize replication service
	replicationService := handleWireInjector(wire.InitializeReplicationService()).(*replication.Service)

	// Register handlers
	deviceHandler := handleWireInjector(wire.InitializeDeviceHandler()).(*handlers.DeviceHandler)
	tenantHandler := handleWireInjector(wire.InitializeTenantHandler()).(*handlers.TenantHandler)
	taskHandler := handleWireInjector(wire.InitializeTaskHandler()).(*handlers.TaskHandler)
	commandHandler := handleWireInjector(wire.InitializeCommandHandler()).(*handlers.CommandHandler)
	scheduledTaskHandler := handleWireInjector(wire.InitializeScheduledTaskHandler()).(*handlers.ScheduledTaskHandler)
	tenantConfigurationHandler := handleWireInjector(wire.InitializeTenantConfigurationHandler()).(*handlers.TenantConfigurationHandler)
	userHandler := handleWireInjector(wire.InitializeUserHandler()).(*handlers.UserHandler)

	if err := replicationService.RegisterHandler(deviceHandler); err != nil {
		slog.Error("failed to register device handler", slog.Any("error", err))
		panic(err)
	}

	if err := replicationService.RegisterHandler(tenantHandler); err != nil {
		slog.Error("failed to register tenant handler", slog.Any("error", err))
		panic(err)
	}

	if err := replicationService.RegisterHandler(taskHandler); err != nil {
		slog.Error("failed to register task handler", slog.Any("error", err))
		panic(err)
	}

	if err := replicationService.RegisterHandler(commandHandler); err != nil {
		slog.Error("failed to register command handler", slog.Any("error", err))
		panic(err)
	}

	if err := replicationService.RegisterHandler(scheduledTaskHandler); err != nil {
		slog.Error("failed to register scheduled task handler", slog.Any("error", err))
		panic(err)
	}

	if err := replicationService.RegisterHandler(tenantConfigurationHandler); err != nil {
		slog.Error("failed to register tenant configuration handler", slog.Any("error", err))
		panic(err)
	}

	if err := replicationService.RegisterHandler(userHandler); err != nil {
		slog.Error("failed to register user handler", slog.Any("error", err))
		panic(err)
	}

	// Start the replication service
	if err := replicationService.Start(); err != nil {
		slog.Error("failed to start replication service", slog.Any("error", err))
		panic(err)
	}

	slog.Info("replication service started successfully")
	return replicationService
}

func slogReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.File = filepath.Base(source.File)
		return slog.Any(a.Key, source)
	}
	return a
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

	traceShutdownFunc, err := startTraceProvider(ctx)
	if err != nil {
		return nil, err
	}

	return func() error {
		if err := metricsShutdownFunc(); err != nil {
			return err
		}
		if err := traceShutdownFunc(); err != nil {
			return err
		}
		return nil
	}, nil
}

func startTraceProvider(ctx context.Context) (ShutdownFunc, error) {
	exp, err := newTraceExporter(ctx)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("zensor-server"),
		)),
	)
	otel.SetTracerProvider(tp)

	return func() error {
		return tp.Shutdown(ctx)
	}, nil
}

func newTraceExporter(ctx context.Context) (trace.SpanExporter, error) {
	endpoint := _defautlEndpoint
	if value, ok := os.LookupEnv("ZENSOR_SERVER_OTELCOL_ENDPOINT"); ok {
		endpoint = value
	}

	return otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
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
