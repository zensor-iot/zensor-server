//go:build wireinject
// +build wireinject

package wire

import (
	"os"
	"time"
	"zensor-server/cmd/config"
	"zensor-server/internal/control_plane/communication"
	"zensor-server/internal/control_plane/httpapi"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/pubsub"
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
		communication.NewCommandPublisher,
		wire.Bind(new(usecases.CommandPublisher), new(*communication.CommandPublisher)),
		persistence.NewDeviceRepository,
		wire.Bind(new(usecases.DeviceRepository), new(*persistence.SimpleDeviceRepository)),
		wire.Bind(new(usecases.CommandRepository), new(*persistence.SimpleDeviceRepository)),
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
		provideKafkaPublisherFactoryOptions,
		pubsub.NewKafkaPublisherFactory,
		wire.Bind(new(pubsub.PublisherFactory), new(*pubsub.KafkaPublisherFactory)),
		provideDatabase,
		persistence.NewScheduledTaskRepository,
		wire.Bind(new(usecases.ScheduledTaskRepository), new(*persistence.SimpleScheduledTaskRepository)),
		persistence.NewDeviceRepository,
		wire.Bind(new(usecases.DeviceRepository), new(*persistence.SimpleDeviceRepository)),
		communication.NewCommandPublisher,
		wire.Bind(new(usecases.CommandPublisher), new(*communication.CommandPublisher)),
		usecases.NewDeviceService,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
		persistence.NewTenantRepository,
		wire.Bind(new(usecases.TenantRepository), new(*persistence.SimpleTenantRepository)),
		usecases.NewTenantService,
		wire.Bind(new(usecases.TenantService), new(*usecases.SimpleTenantService)),
		usecases.NewScheduledTaskService,
		wire.Bind(new(usecases.ScheduledTaskService), new(*usecases.SimpleScheduledTaskService)),
		httpapi.NewScheduledTaskController,
	)

	return nil, nil
}

func InitializeScheduledTaskWorker(broker async.InternalBroker) (*usecases.ScheduledTaskWorker, error) {
	wire.Build(
		provideAppConfig,
		provideTicker,
		persistence.NewScheduledTaskRepository,
		wire.Bind(new(usecases.ScheduledTaskRepository), new(*persistence.SimpleScheduledTaskRepository)),
		persistence.NewTaskRepository,
		wire.Bind(new(usecases.TaskRepository), new(*persistence.SimpleTaskRepository)),
		usecases.NewTaskService,
		wire.Bind(new(usecases.TaskService), new(*usecases.SimpleTaskService)),
		wire.Bind(new(usecases.CommandRepository), new(*persistence.SimpleDeviceRepository)),
		DeviceServiceSet,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
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

func InitializeDeviceService() (usecases.DeviceService, error) {
	wire.Build(
		provideAppConfig,
		DeviceServiceSet,
		wire.Bind(new(usecases.DeviceService), new(*usecases.SimpleDeviceService)),
	)

	return nil, nil
}

var DeviceServiceSet = wire.NewSet(
	provideDatabase,
	provideKafkaPublisherFactoryOptions,
	pubsub.NewKafkaPublisherFactory,
	wire.Bind(new(pubsub.PublisherFactory), new(*pubsub.KafkaPublisherFactory)),
	persistence.NewDeviceRepository,
	wire.Bind(new(usecases.DeviceRepository), new(*persistence.SimpleDeviceRepository)),
	communication.NewCommandPublisher,
	wire.Bind(new(usecases.CommandPublisher), new(*communication.CommandPublisher)),
	usecases.NewDeviceService,
)

func providePubSubFactory(config config.AppConfig) *pubsub.Factory {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "production"
	}

	return pubsub.NewFactory(pubsub.FactoryOptions{
		Environment:   env,
		KafkaBrokers:  config.Kafka.Brokers,
		ConsumerGroup: "zensor-server",
	})
}

func providePublisherFactory(factory *pubsub.Factory) pubsub.PublisherFactory {
	return factory.GetPublisherFactory()
}

func provideKafkaPublisherFactoryOptions(config config.AppConfig) pubsub.KafkaPublisherFactoryOptions {
	return pubsub.KafkaPublisherFactoryOptions{
		Brokers: config.Kafka.Brokers,
	}
}

func provideDatabase(config config.AppConfig) sql.ORM {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "production"
	}

	if env == "local" {
		orm, err := sql.NewMemoryORM("migrations", config.Postgresql.MigrationReplacements)
		if err != nil {
			panic(err)
		}

		return orm
	}

	db := sql.NewPosgreDatabase(config.Postgresql.URL)
	if err := db.Open(); err != nil {
		panic(err)
	}

	db.Up("migrations", config.Postgresql.MigrationReplacements)

	orm, err := sql.NewPosgreORM(config.Postgresql.DSN)
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
		provideKafkaPublisherFactoryOptions,
		pubsub.NewKafkaPublisherFactory,
		wire.Bind(new(pubsub.PublisherFactory), new(*pubsub.KafkaPublisherFactory)),
		communication.NewCommandPublisher,
		wire.Bind(new(usecases.CommandPublisher), new(*communication.CommandPublisher)),
		persistence.NewDeviceRepository,
		wire.Bind(new(usecases.CommandRepository), new(*persistence.SimpleDeviceRepository)),
		usecases.NewCommandWorker,
	)
	return nil, nil
}

func provideTicker() *time.Ticker {
	ticker := time.NewTicker(30 * time.Second)
	return ticker
}

func InitializeDeviceMessageWebSocketController(broker async.InternalBroker) (*httpapi.DeviceMessageWebSocketController, error) {
	wire.Build(
		httpapi.NewDeviceMessageWebSocketController,
	)
	return nil, nil
}
