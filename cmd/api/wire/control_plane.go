//go:build wireinject
// +build wireinject

package wire

import (
	"log/slog"
	"os"
	"sync"
	"time"
	"zensor-server/cmd/config"
	"zensor-server/internal/control_plane/httpapi"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/workers"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/cache"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/notification"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/replication"
	"zensor-server/internal/infra/replication/handlers"
	"zensor-server/internal/infra/sql"

	"github.com/google/wire"
)

func InitializeEvaluationRuleController() (*httpapi.EvaluationRuleController, error) {
	wire.Build(
		provideAppConfig,
		persistence.NewEvaluationRuleRepository,
		wire.Bind(new(usecases.EvaluationRuleRepository), new(*persistence.EvaluationRuleRepository)),
		DeviceServiceSet,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
		usecases.NewEvaluationRuleService,
		wire.Bind(new(usecases.EvaluationRuleService), new(*usecases.SimpleEvaluationRuleService)),
		httpapi.NewEvaluationRuleController,
	)
	return nil, nil
}

func InitializeDeviceController() (*httpapi.DeviceController, error) {
	wire.Build(
		provideAppConfig,
		DeviceServiceSet,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
		httpapi.NewDeviceController,
	)

	return nil, nil
}

func InitializeTaskController() (*httpapi.TaskController, error) {
	wire.Build(
		provideAppConfig,
		providePubSubFactory,
		providePublisherFactory,
		persistence.NewTaskRepository,
		wire.Bind(new(usecases.TaskRepository), new(*persistence.SimpleTaskRepository)),
		provideDatabase,
		persistence.NewDeviceRepository,
		wire.Bind(new(usecases.DeviceRepository), new(*persistence.SimpleDeviceRepository)),
		persistence.NewCommandRepository,
		wire.Bind(new(usecases.CommandRepository), new(*persistence.SimpleCommandRepository)),
		usecases.NewTaskService,
		wire.Bind(new(usecases.TaskService), new(*usecases.SimpleTaskService)),
		usecases.NewDeviceService,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
		httpapi.NewTaskController,
	)

	return nil, nil
}

func InitializeScheduledTaskController() (*httpapi.ScheduledTaskController, error) {
	wire.Build(
		provideAppConfig,
		providePublisherFactoryForEnvironment,
		provideDatabase,
		persistence.NewScheduledTaskRepository,
		wire.Bind(new(usecases.ScheduledTaskRepository), new(*persistence.SimpleScheduledTaskRepository)),
		persistence.NewDeviceRepository,
		wire.Bind(new(usecases.DeviceRepository), new(*persistence.SimpleDeviceRepository)),
		usecases.NewDeviceService,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
		persistence.NewTenantRepository,
		wire.Bind(new(usecases.TenantRepository), new(*persistence.SimpleTenantRepository)),
		usecases.NewTenantService,
		wire.Bind(new(usecases.TenantService), new(*usecases.SimpleTenantService)),
		usecases.NewScheduledTaskService,
		wire.Bind(new(usecases.ScheduledTaskService), new(*usecases.SimpleScheduledTaskService)),
		persistence.NewTaskRepository,
		wire.Bind(new(usecases.TaskRepository), new(*persistence.SimpleTaskRepository)),
		persistence.NewCommandRepository,
		wire.Bind(new(usecases.CommandRepository), new(*persistence.SimpleCommandRepository)),
		usecases.NewTaskService,
		wire.Bind(new(usecases.TaskService), new(*usecases.SimpleTaskService)),
		httpapi.NewScheduledTaskController,
	)

	return nil, nil
}

func InitializeScheduledTaskWorker(broker async.InternalBroker) (*usecases.ScheduledTaskWorker, error) {
	wire.Build(
		provideAppConfig,
		provideTicker,
		provideDatabase,
		providePublisherFactoryForEnvironment,
		persistence.NewScheduledTaskRepository,
		wire.Bind(new(usecases.ScheduledTaskRepository), new(*persistence.SimpleScheduledTaskRepository)),
		persistence.NewTaskRepository,
		wire.Bind(new(usecases.TaskRepository), new(*persistence.SimpleTaskRepository)),
		persistence.NewDeviceRepository,
		wire.Bind(new(usecases.DeviceRepository), new(*persistence.SimpleDeviceRepository)),
		persistence.NewCommandRepository,
		wire.Bind(new(usecases.CommandRepository), new(*persistence.SimpleCommandRepository)),
		usecases.NewTaskService,
		wire.Bind(new(usecases.TaskService), new(*usecases.SimpleTaskService)),
		usecases.NewDeviceService,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
		persistence.NewTenantConfigurationRepository,
		wire.Bind(new(usecases.TenantConfigurationRepository), new(*persistence.SimpleTenantConfigurationRepository)),
		usecases.NewTenantConfigurationService,
		wire.Bind(new(usecases.TenantConfigurationService), new(*usecases.SimpleTenantConfigurationService)),
		usecases.NewScheduledTaskWorker,
	)
	return nil, nil
}

func InitializeTenantController() (*httpapi.TenantController, error) {
	wire.Build(
		provideAppConfig,
		persistence.NewTenantRepository,
		wire.Bind(new(usecases.TenantRepository), new(*persistence.SimpleTenantRepository)),
		DeviceServiceSet,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
		usecases.NewTenantService,
		wire.Bind(new(usecases.TenantService), new(*usecases.SimpleTenantService)),
		httpapi.NewTenantController,
	)

	return nil, nil
}

func InitializeTenantConfigurationController() (*httpapi.TenantConfigurationController, error) {
	wire.Build(
		provideAppConfig,
		providePublisherFactoryForEnvironment,
		provideDatabase,
		persistence.NewTenantConfigurationRepository,
		wire.Bind(new(usecases.TenantConfigurationRepository), new(*persistence.SimpleTenantConfigurationRepository)),
		usecases.NewTenantConfigurationService,
		wire.Bind(new(usecases.TenantConfigurationService), new(*usecases.SimpleTenantConfigurationService)),
		httpapi.NewTenantConfigurationController,
	)

	return nil, nil
}

func InitializeDeviceService() (usecases.DeviceService, error) {
	wire.Build(
		provideAppConfig,
		DeviceServiceSet,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
	)

	return nil, nil
}

func InitializeLoraIntegrationWorker(ticker *time.Ticker, mqttClient mqtt.Client, broker async.InternalBroker, consumerFactory pubsub.ConsumerFactory) (*workers.LoraIntegrationWorker, error) {
	wire.Build(
		provideAppConfig,
		DeviceServiceSet,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
		provideDeviceStateCacheService,
		workers.NewLoraIntegrationWorker,
	)
	return nil, nil
}

var DeviceServiceSet = wire.NewSet(
	provideDatabase,
	persistence.NewDeviceRepository,
	wire.Bind(new(usecases.DeviceRepository), new(*persistence.SimpleDeviceRepository)),
	providePublisherFactoryForEnvironment,
	persistence.NewCommandRepository,
	wire.Bind(new(usecases.CommandRepository), new(*persistence.SimpleCommandRepository)),
	usecases.NewDeviceService,
)

func providePubSubFactory(config config.AppConfig) *pubsub.Factory {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "production"
	}

	return pubsub.NewFactory(pubsub.FactoryOptions{
		Environment:       env,
		KafkaBrokers:      config.Kafka.Brokers,
		ConsumerGroup:     "zensor-server",
		SchemaRegistryURL: config.Kafka.SchemaRegistry,
	})
}

func providePublisherFactory(factory *pubsub.Factory) pubsub.PublisherFactory {
	return factory.GetPublisherFactory()
}

func provideKafkaPublisherFactoryOptions(config config.AppConfig) pubsub.KafkaPublisherFactoryOptions {
	return pubsub.KafkaPublisherFactoryOptions{
		Brokers:           config.Kafka.Brokers,
		SchemaRegistryURL: config.Kafka.SchemaRegistry,
	}
}

func provideDatabase(config config.AppConfig) sql.ORM {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "production"
	}

	if env == "local" {
		orm, err := sql.NewMemoryORM("migrations")
		if err != nil {
			panic(err)
		}

		return orm
	}

	orm, err := sql.NewPosgreORMWithTimeout(config.Postgresql.DSN, config.Postgresql.QueryTimeout)
	if err != nil {
		panic(err)
	}

	return orm
}

func InitializeCommandWorker(broker async.InternalBroker) (*usecases.CommandWorker, error) {
	wire.Build(
		provideAppConfig,
		provideTicker,
		provideDatabase,
		providePublisherFactoryForEnvironment,
		persistence.NewCommandRepository,
		wire.Bind(new(usecases.CommandRepository), new(*persistence.SimpleCommandRepository)),
		usecases.NewCommandWorker,
	)
	return nil, nil
}

func InitializeNotificationWorker(broker async.InternalBroker) (*usecases.NotificationWorker, error) {
	wire.Build(
		provideAppConfig,
		provideTicker,
		provideNotificationClient,
		DeviceServiceSet,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
		persistence.NewTenantConfigurationRepository,
		wire.Bind(new(usecases.TenantConfigurationRepository), new(*persistence.SimpleTenantConfigurationRepository)),
		usecases.NewTenantConfigurationService,
		wire.Bind(new(usecases.TenantConfigurationService), new(*usecases.SimpleTenantConfigurationService)),
		usecases.NewNotificationWorker,
	)
	return nil, nil
}

func provideTicker() *time.Ticker {
	ticker := time.NewTicker(30 * time.Second)
	return ticker
}

func provideNotificationClient(config config.AppConfig) notification.NotificationClient {
	mailerSendConfig := notification.MailerSendConfig{
		APIKey:    config.MailerSend.APIKey,
		FromEmail: config.MailerSend.FromEmail,
		FromName:  config.MailerSend.FromName,
	}

	return notification.NewMailerSendClient(mailerSendConfig)
}

func InitializeDeviceMessageWebSocketController(broker async.InternalBroker) (*httpapi.DeviceMessageWebSocketController, error) {
	wire.Build(
		provideDeviceStateCacheService,
		httpapi.NewDeviceMessageWebSocketController,
	)
	return nil, nil
}

func InitializeDeviceSpecificWebSocketController(broker async.InternalBroker) (*httpapi.DeviceSpecificWebSocketController, error) {
	wire.Build(
		provideDeviceStateCacheService,
		httpapi.NewDeviceSpecificWebSocketController,
	)
	return nil, nil
}

func InitializeReplicationService() (*replication.Service, error) {
	wire.Build(
		provideAppConfig,
		provideMemoryConsumerFactory,
		provideDatabase,
		replication.NewService,
	)
	return nil, nil
}

func provideMemoryConsumerFactory() pubsub.ConsumerFactory {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "production"
	}

	if env == "local" {
		return pubsub.NewMemoryConsumerFactory("replicator")
	}

	return nil
}

func InitializeDeviceHandler() (*handlers.DeviceHandler, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		handlers.NewDeviceHandler,
	)
	return nil, nil
}

func InitializeTenantHandler() (*handlers.TenantHandler, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		handlers.NewTenantHandler,
	)
	return nil, nil
}

func InitializeTaskHandler() (*handlers.TaskHandler, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		handlers.NewTaskHandler,
	)
	return nil, nil
}

func InitializeCommandHandler() (*handlers.CommandHandler, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		handlers.NewCommandHandler,
	)
	return nil, nil
}

func InitializeScheduledTaskHandler() (*handlers.ScheduledTaskHandler, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		handlers.NewScheduledTaskHandler,
	)
	return nil, nil
}

func providePublisherFactoryForEnvironment(config config.AppConfig) pubsub.PublisherFactory {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "production"
	}

	if env == "local" {
		return pubsub.NewMemoryPublisherFactory()
	}

	kafkaOptions := provideKafkaPublisherFactoryOptions(config)
	return pubsub.NewKafkaPublisherFactory(kafkaOptions)
}

var (
	deviceStateCacheService usecases.DeviceStateCacheService
	deviceStateCacheOnce    sync.Once
)

func provideDeviceStateCacheService() usecases.DeviceStateCacheService {
	deviceStateCacheOnce.Do(func() {
		appConfig := provideAppConfig()

		redisCache, err := cache.NewRedisCache(&cache.RedisConfig{
			Addr:     appConfig.Redis.Addr,
			Password: appConfig.Redis.Password,
			DB:       appConfig.Redis.DB,
		})
		if err != nil {
			slog.Error("failed to create Redis cache", slog.String("error", err.Error()))
			deviceStateCacheService = persistence.NewSimpleDeviceStateCacheService()
			slog.Info("falling back to simple device state cache service")
			return
		}

		deviceStateCacheService, err = persistence.NewRedisDeviceStateCacheService(&persistence.RedisDeviceStateCacheConfig{
			Cache:      redisCache,
			KeyPrefix:  "device_state:",
			DefaultTTL: 24 * time.Hour,
		})
		if err != nil {
			slog.Error("failed to create Redis device state cache service", slog.String("error", err.Error()))
			deviceStateCacheService = persistence.NewSimpleDeviceStateCacheService()
			slog.Info("falling back to simple device state cache service")
			return
		}

		slog.Info("Redis device state cache service singleton created")
	})
	return deviceStateCacheService
}
