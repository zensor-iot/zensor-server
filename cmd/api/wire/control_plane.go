//go:build wireinject
// +build wireinject

package wire

import (
	"log/slog"
	"os"
	"sync"
	"time"
	"zensor-server/internal/control_plane/httpapi"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/workers"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/cache"
	"zensor-server/internal/infra/config"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/notification"
	"zensor-server/internal/infra/sql"
	sharedPersistence "zensor-server/internal/shared_kernel/persistence"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"

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
		provideDatabase,
		persistence.NewScheduledTaskRepository,
		wire.Bind(new(usecases.ScheduledTaskRepository), new(*persistence.SimpleScheduledTaskRepository)),
		persistence.NewDeviceRepository,
		wire.Bind(new(usecases.DeviceRepository), new(*persistence.SimpleDeviceRepository)),
		usecases.NewDeviceService,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		wire.Bind(new(sharedUsecases.DeviceAdopter), new(*usecases.SimpleDeviceService)),
		sharedUsecases.NewTenantService,
		wire.Bind(new(sharedUsecases.TenantService), new(*sharedUsecases.SimpleTenantService)),
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
		sharedPersistence.NewTenantConfigurationRepository,
		wire.Bind(new(sharedUsecases.TenantConfigurationRepository), new(*sharedPersistence.SimpleTenantConfigurationRepository)),
		sharedPersistence.NewUserRepository,
		wire.Bind(new(sharedUsecases.UserRepository), new(*sharedPersistence.SimpleUserRepository)),
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		sharedUsecases.NewUserService,
		wire.Bind(new(sharedUsecases.UserService), new(*sharedUsecases.SimpleUserService)),
		sharedUsecases.NewTenantConfigurationService,
		wire.Bind(new(sharedUsecases.TenantConfigurationService), new(*sharedUsecases.SimpleTenantConfigurationService)),
		usecases.NewScheduledTaskWorker,
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

func InitializeLoraIntegrationWorker(ticker *time.Ticker, mqttClient mqtt.Client, broker async.InternalBroker) (*workers.LoraIntegrationWorker, error) {
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
	persistence.NewCommandRepository,
	wire.Bind(new(usecases.CommandRepository), new(*persistence.SimpleCommandRepository)),
	usecases.NewDeviceService,
)

func provideAppConfig() config.AppConfig {
	return config.LoadConfig()
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
		persistence.NewTaskRepository,
		wire.Bind(new(usecases.TaskRepository), new(*persistence.SimpleTaskRepository)),
		usecases.NewTaskService,
		wire.Bind(new(usecases.TaskService), new(*usecases.SimpleTaskService)),
		sharedPersistence.NewTenantConfigurationRepository,
		wire.Bind(new(sharedUsecases.TenantConfigurationRepository), new(*sharedPersistence.SimpleTenantConfigurationRepository)),
		sharedPersistence.NewUserRepository,
		wire.Bind(new(sharedUsecases.UserRepository), new(*sharedPersistence.SimpleUserRepository)),
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		sharedUsecases.NewUserService,
		wire.Bind(new(sharedUsecases.UserService), new(*sharedUsecases.SimpleUserService)),
		sharedUsecases.NewTenantConfigurationService,
		wire.Bind(new(sharedUsecases.TenantConfigurationService), new(*sharedUsecases.SimpleTenantConfigurationService)),
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

func InitializeMetricWorkerFactory(broker async.InternalBroker) *usecases.MetricWorkerFactory {
	return usecases.NewMetricWorkerFactory(broker)
}
