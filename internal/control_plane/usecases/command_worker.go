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
	"go.opentelemetry.io/otel/trace"
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
	}
}

type CommandWorker struct {
	ticker            *time.Ticker
	commandRepository CommandRepository
	broker            async.InternalBroker
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
	for {
		select {
		case <-ctx.Done():
			slog.Info("command worker cancelled")
			wg.Wait()
			return
		case msg := <-subscription.Receiver:
			switch msg.Event {
			case "command_sent":
				wg.Add(1)
				procCtx := context.Background()
				w.handleCommandSent(procCtx, msg.Value.(device.Command), wg.Done)
			case "command_status_update":
				wg.Add(1)
				procCtx := context.Background()
				cmdStatusUpdate, ok := msg.Value.(domain.CommandStatusUpdate)
				if !ok {
					slog.Error("failed to cast command status update data",
						slog.String("type", fmt.Sprintf("%T", msg.Value)),
						slog.String("expected", "domain.CommandStatusUpdate"))
					return
				}
				w.handleCommandStatusUpdate(procCtx, cmdStatusUpdate, wg.Done)
			default:
				slog.Warn("event not supported", slog.String("event", msg.Event))
			}
		case <-w.ticker.C:
			wg.Add(1)
			tickCtx := context.Background()
			tickCtx, _ = otel.Tracer("zensor_server").Start(tickCtx, "reconciliation")
			w.reconciliation(tickCtx, wg.Done)
		}
	}
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
	slog.Info("reconciliation did end",
		slog.Int("commands_processed", len(commands)),
		slog.Time("time", time.Now()),
	)
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

	brokerMsg := async.BrokerMessage{
		Event: "command_processed",
		Value: cmd,
	}
	if err := w.broker.Publish(ctx, async.BrokerTopicName("command_events"), brokerMsg); err != nil {
		slog.Error("failed to publish command processed event", slog.Any("error", err))
	}
}

func (w *CommandWorker) handleCommandSent(ctx context.Context, cmd device.Command, done func()) {
	defer done()

	existingCmd, err := w.commandRepository.GetByID(ctx, domain.ID(cmd.ID))
	if err != nil {
		slog.Error("failed to get command by ID", slog.Any("error", err))
		return
	}

	existingCmd.Sent = true
	existingCmd.SentAt = utils.Time{Time: time.Now()}
	err = w.commandRepository.Update(ctx, existingCmd)
	if err != nil {
		slog.Error("failed to update command", slog.Any("error", err))
	}

	// Metrics are now handled by MetricPublisherWorker
}

func (w *CommandWorker) handleCommandStatusUpdate(ctx context.Context, statusUpdate domain.CommandStatusUpdate, done func()) {
	defer done()
	slog.Debug("received command status update",
		slog.String("command_id", statusUpdate.CommandID),
		slog.String("device_name", statusUpdate.DeviceName),
		slog.String("status", string(statusUpdate.Status)),
	)

	if statusUpdate.CommandID == "" {
		slog.Error("command status update received without command ID",
			slog.String("device_name", statusUpdate.DeviceName),
			slog.String("status", string(statusUpdate.Status)))
		return
	}

	cmdID := domain.ID(statusUpdate.CommandID)
	targetCommand, err := w.commandRepository.GetByID(ctx, cmdID)
	if err != nil {
		slog.Error("failed to find command by ID",
			slog.String("command_id", statusUpdate.CommandID),
			slog.String("error", err.Error()))
		return
	}

	targetCommand.UpdateStatus(statusUpdate.Status, statusUpdate.ErrorMessage)
	err = w.commandRepository.Update(ctx, targetCommand)
	if err != nil {
		slog.Error("failed to update command status",
			slog.String("command_id", targetCommand.ID.String()),
			slog.String("status", string(statusUpdate.Status)),
			slog.String("error", err.Error()))
		return
	}

	slog.Info("command status updated successfully",
		slog.String("command_id", targetCommand.ID.String()),
		slog.String("status", string(statusUpdate.Status)))
}

func (w *CommandWorker) Shutdown() {
	slog.Warn("shutdown is not yet implemented")
}
