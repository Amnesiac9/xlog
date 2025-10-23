package xlog

import (
	"context"
	"log/slog"
)

// MultiHandler provides an easy way to output logs to multiple handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, lvl) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		rr := r // copy; Handle consumes the record
		if err := h.Handle(ctx, rr); err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	news := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		news[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: news}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	news := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		news[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: news}
}
