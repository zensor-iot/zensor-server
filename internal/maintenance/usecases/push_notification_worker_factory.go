package usecases

import (
	"fmt"
	"log/slog"
	"zensor-server/cmd/config"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/notification"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"
)

type PushNotificationWorkerFactory struct {
	broker             async.InternalBroker
	notificationClient notification.NotificationClient
	pushTokenService   sharedUsecases.PushTokenService
	userService        sharedUsecases.UserService
}

func NewPushNotificationWorkerFactory(
	broker async.InternalBroker,
	notificationClient notification.NotificationClient,
	pushTokenService sharedUsecases.PushTokenService,
	userService sharedUsecases.UserService,
) *PushNotificationWorkerFactory {
	return &PushNotificationWorkerFactory{
		broker:             broker,
		notificationClient: notificationClient,
		pushTokenService:   pushTokenService,
		userService:        userService,
	}
}

func (f *PushNotificationWorkerFactory) CreateWorkers(cfg config.PushNotificationsConfig) ([]*PushNotificationWorker, error) {
	var workers []*PushNotificationWorker

	for _, workerCfg := range cfg {
		slog.Info("creating push notification worker",
			slog.String("name", workerCfg.Name),
			slog.String("topic", workerCfg.Topic),
			slog.String("event_type", workerCfg.EventType))

		worker, err := NewPushNotificationWorker(
			workerCfg,
			f.broker,
			f.notificationClient,
			f.pushTokenService,
			f.userService,
		)
		if err != nil {
			return nil, fmt.Errorf("creating push notification worker %s: %w", workerCfg.Name, err)
		}

		workers = append(workers, worker)
	}

	return workers, nil
}
