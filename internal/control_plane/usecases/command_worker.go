package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"zensor-server/internal/control_plane/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	_metricKeyCommands = "commands"
)

func NewCommandWorker(
	ticker *time.Ticker,
	commandRepository CommandRepository,
	publisher CommandPublisher,
) *CommandWorker {
	return &CommandWorker{
		ticker:            ticker,
		commandRepository: commandRepository,
		publisher:         publisher,
		metricCounters:    make(map[string]metric.Float64Counter),
	}
}

type CommandWorker struct {
	ticker            *time.Ticker
	commandRepository CommandRepository
	publisher         CommandPublisher
	metricCounters    map[string]metric.Float64Counter
}

func (w *CommandWorker) Run(ctx context.Context, done func()) {
	slog.Debug("run with context initialized")
	defer done()
	var wg sync.WaitGroup
	w.setupOtelCounters()
	for {
		select {
		case <-ctx.Done():
			slog.Info("command worker cancelled")
			wg.Wait()
			return
		case <-w.ticker.C:
			wg.Add(1)
			tickCtx := context.Background()
			w.reconciliation(tickCtx, wg.Done)
		}
	}
}

func (w *CommandWorker) setupOtelCounters() {
	meter := otel.Meter("zensor_server")
	commandCounter, _ := meter.Float64Counter(
		fmt.Sprintf("%s.%s", "zensor_server", "commands"),
		metric.WithDescription("zensor_server command counter"),
	)

	w.metricCounters[_metricKeyCommands] = commandCounter
}

func (w *CommandWorker) reconciliation(ctx context.Context, done func()) {
	slog.Debug("reconciliation start...", slog.Time("time", time.Now()))
	defer done()
	commands, err := w.commandRepository.FindAllPending(ctx)
	if err != nil {
		slog.Error("finding all pending commands", slog.Any("error", err))
		return
	}

	for _, cmd := range commands {
		w.handle(ctx, cmd)
	}
	slog.Debug("reconciliation end", slog.Time("time", time.Now()))
}

func (w *CommandWorker) handle(ctx context.Context, cmd domain.Command) {
	if cmd.DispatchAfter.Before(time.Now()) {
		slog.Warn("command is not ready to be sent", slog.Time("dispatch_after", cmd.DispatchAfter.Time))
		return
	}

	cmd.Ready = true
	w.publisher.Dispatch(ctx, cmd)

	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String("zensor_server"),
		attribute.String("device_name", cmd.Device.Name),
	}
	w.metricCounters[_metricKeyCommands].Add(ctx, 1, metric.WithAttributes(attributes...))
}

func (w *CommandWorker) Shutdown() {
	slog.Warn("shutdown is not yet implemented")
}
