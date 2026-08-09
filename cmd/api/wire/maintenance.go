//go:build wireinject
// +build wireinject

package wire

import (
	"time"
	controlPlaneUsecases "zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/config"
	maintenanceHTTPAPI "zensor-server/internal/maintenance/httpapi"
	maintenancePersistence "zensor-server/internal/maintenance/persistence"
	maintenanceUsecases "zensor-server/internal/maintenance/usecases"
	sharedPersistence "zensor-server/internal/shared_kernel/persistence"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"

	"github.com/google/wire"
)

func InitializeMaintenanceActivityController() (*maintenanceHTTPAPI.ActivityController, error) {
	wire.Build(
		provideAppConfig,
		maintenancePersistence.NewActivityRepository,
		wire.Bind(new(maintenanceUsecases.ActivityRepository), new(*maintenancePersistence.SimpleActivityRepository)),
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		DeviceServiceSet,
		wire.Bind(new(sharedUsecases.DeviceAdopter), new(*controlPlaneUsecases.SimpleDeviceService)),
		sharedUsecases.NewTenantService,
		wire.Bind(new(sharedUsecases.TenantService), new(*sharedUsecases.SimpleTenantService)),
		maintenanceUsecases.NewActivityService,
		wire.Bind(new(maintenanceUsecases.ActivityService), new(*maintenanceUsecases.SimpleActivityService)),
		maintenanceHTTPAPI.NewActivityController,
	)
	return nil, nil
}

func InitializeMaintenanceExecutionController() (*maintenanceHTTPAPI.ExecutionController, error) {
	wire.Build(
		provideAppConfig,
		maintenancePersistence.NewExecutionRepository,
		wire.Bind(new(maintenanceUsecases.ExecutionRepository), new(*maintenancePersistence.SimpleExecutionRepository)),
		maintenancePersistence.NewActivityRepository,
		wire.Bind(new(maintenanceUsecases.ActivityRepository), new(*maintenancePersistence.SimpleActivityRepository)),
		maintenanceUsecases.NewExecutionService,
		wire.Bind(new(maintenanceUsecases.ExecutionService), new(*maintenanceUsecases.SimpleExecutionService)),
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		DeviceServiceSet,
		wire.Bind(new(sharedUsecases.DeviceAdopter), new(*controlPlaneUsecases.SimpleDeviceService)),
		sharedUsecases.NewTenantService,
		wire.Bind(new(sharedUsecases.TenantService), new(*sharedUsecases.SimpleTenantService)),
		maintenanceUsecases.NewActivityService,
		wire.Bind(new(maintenanceUsecases.ActivityService), new(*maintenanceUsecases.SimpleActivityService)),
		maintenanceHTTPAPI.NewExecutionController,
	)
	return nil, nil
}

func InitializeExecutionWorker(broker async.InternalBroker) (*maintenanceUsecases.ExecutionWorker, error) {
	wire.Build(
		provideAppConfig,
		provideExecutionWorkerTicker,
		maintenancePersistence.NewActivityRepository,
		wire.Bind(new(maintenanceUsecases.ActivityRepository), new(*maintenancePersistence.SimpleActivityRepository)),
		maintenancePersistence.NewExecutionRepository,
		wire.Bind(new(maintenanceUsecases.ExecutionRepository), new(*maintenancePersistence.SimpleExecutionRepository)),
		maintenanceUsecases.NewExecutionService,
		wire.Bind(new(maintenanceUsecases.ExecutionService), new(*maintenanceUsecases.SimpleExecutionService)),
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		sharedPersistence.NewTenantConfigurationRepository,
		wire.Bind(new(sharedUsecases.TenantConfigurationRepository), new(*sharedPersistence.SimpleTenantConfigurationRepository)),
		sharedPersistence.NewUserRepository,
		wire.Bind(new(sharedUsecases.UserRepository), new(*sharedPersistence.SimpleUserRepository)),
		sharedUsecases.NewUserService,
		wire.Bind(new(sharedUsecases.UserService), new(*sharedUsecases.SimpleUserService)),
		DeviceServiceSet,
		wire.Bind(new(sharedUsecases.DeviceAdopter), new(*controlPlaneUsecases.SimpleDeviceService)),
		sharedUsecases.NewTenantService,
		wire.Bind(new(sharedUsecases.TenantService), new(*sharedUsecases.SimpleTenantService)),
		sharedUsecases.NewTenantConfigurationService,
		wire.Bind(new(sharedUsecases.TenantConfigurationService), new(*sharedUsecases.SimpleTenantConfigurationService)),
		maintenanceUsecases.NewExecutionWorker,
	)
	return nil, nil
}

func InitializePushNotificationWorkerFactory(broker async.InternalBroker) (*maintenanceUsecases.PushNotificationWorkerFactory, error) {
	wire.Build(
		provideAppConfig,
		provideCompositeNotificationClient,
		provideDatabase,
		sharedPersistence.NewPushTokenRepository,
		wire.Bind(new(sharedUsecases.PushTokenRepository), new(*sharedPersistence.SimplePushTokenRepository)),
		sharedUsecases.NewPushTokenService,
		wire.Bind(new(sharedUsecases.PushTokenService), new(*sharedUsecases.SimplePushTokenService)),
		sharedPersistence.NewUserRepository,
		wire.Bind(new(sharedUsecases.UserRepository), new(*sharedPersistence.SimpleUserRepository)),
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		sharedUsecases.NewUserService,
		wire.Bind(new(sharedUsecases.UserService), new(*sharedUsecases.SimpleUserService)),
		maintenanceUsecases.NewPushNotificationWorkerFactory,
	)
	return nil, nil
}

func provideExecutionWorkerTicker(appConfig config.AppConfig) *time.Ticker {
	interval := appConfig.ExecutionWorker.TickerInterval
	if interval == 0 {
		interval = 5 * time.Minute
	}
	return time.NewTicker(interval)
}
