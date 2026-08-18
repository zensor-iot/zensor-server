package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"zensor-server/cmd/api/wire"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/config"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/node"
	"zensor-server/internal/infra/o11y"

	maintenanceUsecases "zensor-server/internal/maintenance/usecases"
	victronHTTPAPI "zensor-server/internal/victron/httpapi"
	victronUsecases "zensor-server/internal/victron/usecases"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var logLevelMapping = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func main() {
	appConfig := config.LoadConfig()

	level := logLevelMapping[appConfig.General.LogLevel]
	baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: level, ReplaceAttr: slogReplaceAttr})
	stdoutHandler := baseHandler.WithAttrs([]slog.Attr{slog.String("version", node.Version)})
	otelHandler := otelslog.NewHandler("zensor-server")
	slog.SetDefault(slog.New(o11y.NewTeeHandler(level, stdoutHandler, otelHandler)))
	slog.Info("🚀 zensor is initializing")
	slog.Debug("config loaded", "data", appConfig)

	shutdownOtel := startOTel()

	// DATA PLANE - Set up broker and dependencies first
	internalBroker := async.NewLocalBroker()

	controllers := []httpserver.Controller{
		asController(handleWireInjector(wire.InitializeDeviceController())),
		asController(handleWireInjector(wire.InitializeEvaluationRuleController())),
		asController(handleWireInjector(wire.InitializeTaskController())),
		asController(handleWireInjector(wire.InitializeTenantController())),
		asController(handleWireInjector(wire.InitializeTenantConfigurationController())),
		asController(handleWireInjector(wire.InitializeScheduledTaskController())),
		asController(handleWireInjector(wire.InitializeUserController())),
		asController(handleWireInjector(wire.InitializePushTokenController())),
		asController(handleWireInjector(wire.InitializeWebPushController())),
		asController(handleWireInjector(wire.InitializeDeviceMessageWebSocketController(internalBroker))),
		asController(handleWireInjector(wire.InitializeDeviceSpecificWebSocketController(internalBroker))),
	}

	if appConfig.Modules.Maintenance.Enabled {
		slog.Info("module enabled and will be wired", slog.String("module", "maintenance"))
		controllers = append(controllers,
			asController(handleWireInjector(wire.InitializeMaintenanceActivityController())),
			asController(handleWireInjector(wire.InitializeMaintenanceExecutionController())),
		)
	}

	if appConfig.Modules.Permaculture.Enabled {
		slog.Info("module enabled and will be wired", slog.String("module", "permaculture"))
	}

	if appConfig.Modules.Victron.Enabled {
		slog.Info("module enabled and will be wired", slog.String("module", "victron"))
		controllers = append(controllers, victronHTTPAPI.NewVictronWebSocketController(internalBroker))
	}

	if appConfig.VictoriaMetrics.BaseURL != "" {
		slog.Info("exposing victoriametrics query API", slog.String("base_url", appConfig.VictoriaMetrics.BaseURL))
		controllers = append(controllers, httpserver.NewMetricsProxyController(appConfig.VictoriaMetrics.BaseURL))
	}

	var httpServer httpserver.Server
	switch {
	case appConfig.Auth.Enabled && appConfig.Auth.Mode == config.AuthModeStatic:
		slog.Warn("authentication enabled in STATIC mode: single hardcoded admin user, do not use in production")
		staticAuthComponents := asComponents[wire.StaticAuthComponents](handleWireInjector(wire.InitializeStaticAuthComponents()))
		apiKeyComponents := asComponents[wire.APIKeyComponents](handleWireInjector(wire.InitializeAPIKeyComponents()))
		controllers = append(controllers, staticAuthComponents.Controller, apiKeyComponents.Controller)
		httpServer = httpserver.NewServerWithAuth(appConfig.HTTP.Port, staticAuthComponents.Service, apiKeyComponents.Service, controllers...)
	case appConfig.Auth.Enabled:
		slog.Info("authentication enabled: session middleware will protect /v1 and /ws routes")
		authComponents := asComponents[wire.AuthComponents](handleWireInjector(wire.InitializeAuthComponents()))
		apiKeyComponents := asComponents[wire.APIKeyComponents](handleWireInjector(wire.InitializeAPIKeyComponents()))
		controllers = append(controllers, authComponents.Controller, apiKeyComponents.Controller)
		httpServer = httpserver.NewServerWithAuth(appConfig.HTTP.Port, authComponents.Service, apiKeyComponents.Service, controllers...)
	default:
		slog.Warn("authentication disabled: trusting X-User headers")
		httpServer = httpserver.NewServer(appConfig.HTTP.Port, controllers...)
	}

	appCtx, cancelFn := context.WithCancel(context.Background())
	go httpServer.Run()

	env, envOK := os.LookupEnv("ENV")
	if !envOK {
		env = "production"
	}

	var wg sync.WaitGroup
	ticker := time.NewTicker(30 * time.Second)
	var mqttClient mqtt.Client
	if env == "local" {
		mqttClient = mqtt.NewNoOpClient()
	} else {
		simpleClientOpts := mqtt.SimpleClientOpts{
			Broker:   appConfig.MQTTClient.Broker,
			ClientID: appConfig.MQTTClient.ClientID,
			Username: appConfig.MQTTClient.Username,
			Password: appConfig.MQTTClient.Password, // pragma: allowlist secret
		}
		mqttClient = mqtt.NewSimpleClient(simpleClientOpts)
	}

	// TODO: capture workers into a variable to shutdown them later
	if appConfig.Modules.Permaculture.Enabled {
		wg.Add(1)
		go asWorker(handleWireInjector(wire.InitializeLoraIntegrationWorker(ticker, mqttClient, internalBroker))).Run(appCtx, wg.Done)
		wg.Add(1)
		go asWorker(handleWireInjector(wire.InitializeCommandWorker(internalBroker))).Run(appCtx, wg.Done)
		wg.Add(1)
		go asWorker(handleWireInjector(wire.InitializeScheduledTaskWorker(internalBroker))).Run(appCtx, wg.Done)
		wg.Add(1)
		go asWorker(handleWireInjector(wire.InitializeNotificationWorker(internalBroker))).Run(appCtx, wg.Done)
	}

	if appConfig.Modules.Maintenance.Enabled {
		wg.Add(1)
		go asWorker(handleWireInjector(wire.InitializeExecutionWorker(internalBroker))).Run(appCtx, wg.Done)

		// Initialize push notification workers based on configuration
		pushNotificationWorkerFactory := asComponents[maintenanceUsecases.PushNotificationWorkerFactory](handleWireInjector(wire.InitializePushNotificationWorkerFactory(internalBroker)))
		pushNotificationWorkers, err := pushNotificationWorkerFactory.CreateWorkers(appConfig.PushNotifications)
		if err != nil {
			slog.Error("failed to create push notification workers", slog.Any("error", err))
			panic(err)
		}

		// Start all push notification workers
		for _, worker := range pushNotificationWorkers {
			wg.Add(1)
			go worker.Run(appCtx, wg.Done)
		}
	}

	// Initialize metric workers based on configuration
	metricWorkerFactory := wire.InitializeMetricWorkerFactory(internalBroker)
	metricWorkers, err := metricWorkerFactory.CreateWorkers(appConfig.Metrics)
	if err != nil {
		slog.Error("failed to create metric workers", slog.Any("error", err))
		panic(err)
	}

	// Start all metric workers
	for _, worker := range metricWorkers {
		wg.Add(1)
		go worker.Run(appCtx, wg.Done)
	}

	if appConfig.Modules.Victron.Enabled {
		slog.Info("module enabled and will be wired", slog.String("module", "victron"))
		var victronMQTTClient mqtt.Client
		if env == "local" {
			victronMQTTClient = mqtt.NewNoOpClient()
		} else {
			victronMQTTClient = mqtt.NewSimpleClient(mqtt.SimpleClientOpts{
				Broker:   appConfig.Victron.MQTT.Broker,
				ClientID: appConfig.Victron.MQTT.ClientID,
				Username: appConfig.Victron.MQTT.Username,
				Password: appConfig.Victron.MQTT.Password,
			})
		}
		wg.Add(1)
		victronWorker := victronUsecases.NewVictronWorker(
			appConfig.Victron.PortalID,
			victronMQTTClient,
			internalBroker,
		)
		go victronWorker.Run(appCtx, wg.Done)

		wg.Add(1)
		victronMetricWorker := victronUsecases.NewVictronMetricWorker(internalBroker)
		go victronMetricWorker.Run(appCtx, wg.Done)
	}

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel
	shutdownOtel()

	cancelFn()
	wg.Wait()
	slog.Info("good bye!!!")
	os.Exit(0)
}

func slogReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		source, ok := a.Value.Any().(*slog.Source)
		if !ok {
			return a
		}
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

var _histogramBuckets = []float64{5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000, 25000, 50000, 100000}

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

	logsShutdownFunc, err := startLogsProvider(ctx)
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
		if err := logsShutdownFunc(); err != nil {
			return err
		}
		return nil
	}, nil
}

func newOTelResource() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("zensor-server"),
		semconv.ServiceVersionKey.String(node.Version),
	)
}

func startLogsProvider(ctx context.Context) (ShutdownFunc, error) {
	exp, err := newLogExporter(ctx)
	if err != nil {
		return nil, err
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)),
		sdklog.WithResource(newOTelResource()),
	)
	global.SetLoggerProvider(lp)

	return func() error {
		return lp.Shutdown(ctx)
	}, nil
}

func newLogExporter(ctx context.Context) (sdklog.Exporter, error) {
	endpoint := _defautlEndpoint
	if value, ok := os.LookupEnv("ZENSOR_SERVER_OTELCOL_ENDPOINT"); ok {
		endpoint = value
	}

	return otlploggrpc.New(
		ctx,
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithInsecure(),
	)
}

func startTraceProvider(ctx context.Context) (ShutdownFunc, error) {
	exp, err := newTraceExporter(ctx)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(newOTelResource()),
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
		metric.WithResource(newOTelResource()),
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

func asController(value any) httpserver.Controller {
	controller, ok := value.(httpserver.Controller)
	if !ok {
		panic(fmt.Sprintf("wire injector did not return an httpserver.Controller, got %T", value))
	}
	return controller
}

func asComponents[T any](value any) *T {
	components, ok := value.(*T)
	if !ok {
		panic(fmt.Sprintf("wire injector did not return %T, got %T", (*T)(nil), value))
	}
	return components
}

func asWorker(value any) async.Worker {
	worker, ok := value.(async.Worker)
	if !ok {
		panic(fmt.Sprintf("wire injector did not return an async.Worker, got %T", value))
	}
	return worker
}
