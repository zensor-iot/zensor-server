package usecases

import (
	"log/slog"

	"zensor-server/cmd/config"
	"zensor-server/internal/infra/async"
)

// MetricWorkerFactory creates metric workers based on configuration
type MetricWorkerFactory struct {
	broker async.InternalBroker
}

// NewMetricWorkerFactory creates a new factory
func NewMetricWorkerFactory(broker async.InternalBroker) *MetricWorkerFactory {
	return &MetricWorkerFactory{
		broker: broker,
	}
}

// CreateWorkers creates all metric workers based on the configuration
func (f *MetricWorkerFactory) CreateWorkers(cfg config.MetricsConfig) ([]*MetricWorker, error) {
	var workers []*MetricWorker

	for _, workerCfg := range cfg {
		slog.Info("creating metric worker",
			slog.String("name", workerCfg.Name),
			slog.String("type", workerCfg.Type),
			slog.String("topic", workerCfg.Topic))

		worker, err := NewMetricWorker(workerCfg, f.broker)
		if err != nil {
			slog.Error("failed to create metric worker",
				slog.String("name", workerCfg.Name),
				slog.Any("error", err))
			return nil, err
		}

		workers = append(workers, worker)
	}

	return workers, nil
}
