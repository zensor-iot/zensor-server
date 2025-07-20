package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/device"
	"zensor-server/internal/shared_kernel/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	_metricKeyCommands = "commands"
)

func NewCommandWorker(
	ticker *time.Ticker,
	commandRepository CommandRepository,
	broker async.InternalBroker,
) *CommandWorker {
	return &CommandWorker{
		ticker:            ticker,
		commandRepository: commandRepository,
		broker:            broker,
		metricCounters:    make(map[string]metric.Float64Counter),
	}
}

type CommandWorker struct {
	ticker            *time.Ticker
	commandRepository CommandRepository
	broker            async.InternalBroker
	metricCounters    map[string]metric.Float64Counter
}

func (w *CommandWorker) Run(ctx context.Context, done func()) {
	slog.Debug("run with context initialized")
	defer done()
	subscription, err := w.broker.Subscribe(async.BrokerTopicName("device_messages"))
	if err != nil {
		slog.Error("subscribing to topic", slog.Any("error", err))
		return
	}
	var wg sync.WaitGroup
	w.setupOtelCounters()
	for {
		select {
		case <-ctx.Done():
			slog.Info("command worker cancelled")
			wg.Wait()
			return
		case msg := <-subscription.Receiver:
			if msg.Event != "command_sent" {
				slog.Warn("event not supported", slog.String("event", msg.Event))
				break
			}
			wg.Add(1)
			procCtx := context.Background()
			w.handleCommandSent(procCtx, msg.Value.(device.Command), wg.Done)
		case <-w.ticker.C:
			wg.Add(1)
			tickCtx := context.Background()
			tickCtx, _ = otel.Tracer("zensor_server").Start(tickCtx, "reconciliation")
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
	span := trace.SpanFromContext(ctx)
	slog.Debug("reconciliation start...", slog.Time("time", time.Now()))
	defer done()
	commands, err := w.commandRepository.FindAllPending(ctx)
	if err != nil {
		slog.Error("finding all pending commands",
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
			slog.Any("error", err),
		)
		return
	}

	for _, cmd := range commands {
		w.handle(ctx, cmd)
	}
	slog.Debug("reconciliation end", slog.Time("time", time.Now()))
}

func (w *CommandWorker) handle(ctx context.Context, cmd domain.Command) {
	span := trace.SpanFromContext(ctx)
	if cmd.DispatchAfter.After(time.Now()) {
		slog.Warn("command is not ready to be sent",
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
			slog.Time("dispatch_after", cmd.DispatchAfter.Time),
		)
		return
	}

	cmd.Ready = true
	err := w.commandRepository.Update(ctx, cmd)
	if err != nil {
		slog.Error("failed to update command",
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
			slog.Any("error", err),
		)
		return
	}

	slog.Debug("new message ready to be sent", slog.String("id", cmd.ID.String()))

	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String("zensor_server"),
		attribute.String("device_name", cmd.Device.Name),
	}
	w.metricCounters[_metricKeyCommands].Add(ctx, 1, metric.WithAttributes(attributes...))
}

func (w *CommandWorker) handleCommandSent(ctx context.Context, cmd device.Command, done func()) {
	defer done()
	domainCmd := domain.Command{
		ID: domain.ID(cmd.ID),
		Device: domain.Device{
			ID:   domain.ID(cmd.DeviceID),
			Name: cmd.DeviceName,
		},
		Task: domain.Task{
			ID: domain.ID(cmd.TaskID),
		},
		Port:     domain.Port(cmd.Port),
		Priority: domain.CommandPriority(cmd.Priority),
		Payload: domain.CommandPayload{
			Index: domain.Index(cmd.Payload.Index),
			Value: domain.CommandValue(cmd.Payload.Value),
		},
		DispatchAfter: cmd.DispatchAfter,
		Ready:         cmd.Ready,
		Sent:          true,
		SentAt:        utils.Time{Time: time.Now()},
	}

	err := w.commandRepository.Update(ctx, domainCmd)
	if err != nil {
		slog.Warn("failed to update command", slog.Any("error", err))
	}

	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String("zensor_server"),
		attribute.String("device_name", cmd.DeviceName),
	}
	w.metricCounters[_metricKeyCommands].Add(ctx, 1, metric.WithAttributes(attributes...))
}

func (w *CommandWorker) Shutdown() {
	slog.Warn("shutdown is not yet implemented")
}
