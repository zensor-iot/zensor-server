//go:build wireinject
// +build wireinject

package wire

import (
	"context"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/config"
	"zensor-server/internal/infra/notification"

	sharedHTTPAPI "zensor-server/internal/shared_kernel/httpapi"
	sharedPersistence "zensor-server/internal/shared_kernel/persistence"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"

	"github.com/google/wire"
)

func InitializeUserController() (*sharedHTTPAPI.UserController, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
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
		provideCompositeNotificationClient,
		sharedPersistence.NewPushTokenRepository,
		wire.Bind(new(sharedUsecases.PushTokenRepository), new(*sharedPersistence.SimplePushTokenRepository)),
		sharedUsecases.NewPushTokenService,
		wire.Bind(new(sharedUsecases.PushTokenService), new(*sharedUsecases.SimplePushTokenService)),
		sharedUsecases.NewUserPushMessageSender,
		wire.Bind(new(sharedUsecases.UserPushMessageSender), new(*sharedUsecases.SimpleUserPushMessageSender)),
		sharedHTTPAPI.NewPushTokenController,
	)
	return nil, nil
}

func provideCompositeNotificationClient(cfg config.AppConfig) notification.NotificationClient {
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
	fcmClient := notification.NewFCMPushClient(context.Background(), fcmConfig)

	webPushClient := notification.NewWebPushClient(notification.WebPushConfig{
		VAPIDPublicKey:  cfg.WebPush.VAPIDPublicKey,
		VAPIDPrivateKey: cfg.WebPush.VAPIDPrivateKey,
		Subscriber:      cfg.WebPush.Subscriber,
	})

	pushClient := notification.NewPlatformRoutingPushClient(fcmClient, webPushClient)

	return notification.NewCompositeNotificationClient(emailClient, pushClient)
}

func provideWebPushController(cfg config.AppConfig) *sharedHTTPAPI.WebPushController {
	return sharedHTTPAPI.NewWebPushController(cfg.WebPush.VAPIDPublicKey)
}

func InitializeWebPushController() (*sharedHTTPAPI.WebPushController, error) {
	wire.Build(
		provideAppConfig,
		provideWebPushController,
	)
	return nil, nil
}
