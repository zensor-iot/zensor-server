//go:build wireinject
// +build wireinject

package wire

import (
	"time"
	"zensor-server/cmd/config"
	"zensor-server/internal/control_plane/communication"
	"zensor-server/internal/control_plane/httpapi"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"

	"github.com/google/wire"
)

func InitializeEvaluationRuleController() (*httpapi.EvaluationRuleController, error) {
	wire.Build(
		provideAppConfig,
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

func provideKafkaPublisherFactoryOptions(config config.AppConfig) pubsub.KafkaPublisherFactoryOptions {
	return pubsub.KafkaPublisherFactoryOptions{
		Brokers: config.Kafka.Brokers,
	}
}

func provideDatabase(config config.AppConfig) sql.ORM {
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

func InitializeCommandWorker() (*usecases.CommandWorker, error) {
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
