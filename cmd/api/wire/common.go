//go:build wireinject
// +build wireinject

package wire

import (
	"context"
	"zensor-server/cmd/config"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/notification"
	"zensor-server/internal/infra/replication/handlers"
	sharedHTTPAPI "zensor-server/internal/shared_kernel/httpapi"
	sharedPersistence "zensor-server/internal/shared_kernel/persistence"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"

	"github.com/google/wire"
)

func InitializeTenantConfigurationHandler() (*handlers.TenantConfigurationHandler, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		handlers.NewTenantConfigurationHandler,
	)
	return nil, nil
}

func InitializeUserHandler() (*handlers.UserHandler, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		handlers.NewUserHandler,
	)
	return nil, nil
}

func InitializeUserController() (*sharedHTTPAPI.UserController, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		providePublisherFactoryForEnvironment,
		sharedPersistence.NewUserRepository,
		wire.Bind(new(sharedUsecases.UserRepository), new(*sharedPersistence.SimpleUserRepository)),
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		sharedUsecases.NewUserService,
		wire.Bind(new(sharedUsecases.UserService), new(*sharedUsecases.SimpleUserService)),
		sharedHTTPAPI.NewUserController,
	)
	return nil, nil
}

func InitializeTenantController() (*sharedHTTPAPI.TenantController, error) {
	wire.Build(
		provideAppConfig,
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		DeviceServiceSet,
		wire.Bind(new(sharedUsecases.DeviceAdopter), new(*usecases.SimpleDeviceService)),
		sharedUsecases.NewTenantService,
		wire.Bind(new(sharedUsecases.TenantService), new(*sharedUsecases.SimpleTenantService)),
		sharedHTTPAPI.NewTenantController,
	)
	return nil, nil
}

func InitializeTenantConfigurationController() (*sharedHTTPAPI.TenantConfigurationController, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		providePublisherFactoryForEnvironment,
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
		sharedHTTPAPI.NewTenantConfigurationController,
	)
	return nil, nil
}

func InitializePushTokenController() (*sharedHTTPAPI.PushTokenController, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		sharedPersistence.NewPushTokenRepository,
		wire.Bind(new(sharedUsecases.PushTokenRepository), new(*sharedPersistence.SimplePushTokenRepository)),
		sharedUsecases.NewPushTokenService,
		wire.Bind(new(sharedUsecases.PushTokenService), new(*sharedUsecases.SimplePushTokenService)),
		sharedHTTPAPI.NewPushTokenController,
	)
	return nil, nil
}

func provideCompositeNotificationClient(cfg config.AppConfig) (notification.NotificationClient, error) {
	mailerSendConfig := notification.MailerSendConfig{
		APIKey:    cfg.MailerSend.APIKey,
		FromEmail: cfg.MailerSend.FromEmail,
		FromName:  cfg.MailerSend.FromName,
	}
	emailClient := notification.NewMailerSendClient(mailerSendConfig)

	fcmConfig := notification.FCMConfig{
		ProjectID:          cfg.FCM.ProjectID,
		ServiceAccountPath: cfg.FCM.ServiceAccountPath,
	}
	pushClient, err := notification.NewFCMClient(context.Background(), fcmConfig)
	if err != nil {
		return nil, err
	}

	return notification.NewCompositeNotificationClient(emailClient, pushClient), nil
}
