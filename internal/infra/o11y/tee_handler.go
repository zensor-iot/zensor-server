// Package o11y provides observability helpers.
package o11y

import (
	"context"
	"log/slog"
)

// TeeHandler is a slog.Handler that fans every record out to multiple
// underlying handlers (e.g. stdout and an OTLP exporter), gated by a
// single minimum level.
type TeeHandler struct {
	level    slog.Leveler
	handlers []slog.Handler
}

var _ slog.Handler = (*TeeHandler)(nil)

func NewTeeHandler(level slog.Leveler, handlers ...slog.Handler) *TeeHandler {
	return &TeeHandler{level: level, handlers: handlers}
}

func (t *TeeHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= t.level.Level()
}

func (t *TeeHandler) Handle(ctx context.Context, record slog.Record) error {
	var firstErr error
	for _, handler := range t.handlers {
		if err := handler.Handle(ctx, record.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (t *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(t.handlers))
	for i, handler := range t.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &TeeHandler{level: t.level, handlers: handlers}
}

func (t *TeeHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(t.handlers))
	for i, handler := range t.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &TeeHandler{level: t.level, handlers: handlers}
}
