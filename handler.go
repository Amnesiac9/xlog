package xlog

import (
	"context"
	"log/slog"
)

type XlogHandler struct {
	handler slog.Handler

	// function to add specific attributes/fields from a given context
	attrFromContext []func(context.Context) []slog.Attr
}

func NewHandler(handler slog.Handler, attrFromContextFuncs ...func(context.Context) []slog.Attr) *XlogHandler {
	return &XlogHandler{
		handler,
		attrFromContextFuncs,
	}
}

var _ slog.Handler = (*XlogHandler)(nil)

func (h *XlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *XlogHandler) Handle(ctx context.Context, rec slog.Record) error {
	// Create a new record so we can safely append attrs
	nr := slog.NewRecord(rec.Time, rec.Level, rec.Message, rec.PC)

	// 1) original record attrs
	rec.Attrs(func(a slog.Attr) bool {
		nr.AddAttrs(a)
		return true
	})

	// 2) attrs extracted from context
	for _, fn := range h.attrFromContext {
		if fn == nil {
			continue
		}
		for _, a := range fn(ctx) {
			nr.AddAttrs(a)
		}
	}

	// Pass through to the wrapped handler; ReplaceAttr/AddSource apply there.
	return h.handler.Handle(ctx, nr)
}

func (h *XlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &XlogHandler{
		handler:         h.handler.WithAttrs(attrs),
		attrFromContext: h.attrFromContext,
	}
}

func (h *XlogHandler) WithGroup(name string) slog.Handler {
	return &XlogHandler{
		handler:         h.handler.WithGroup(name),
		attrFromContext: h.attrFromContext,
	}
}

// // Returns a func that extracts any given keys from the context
// func ExtractFromContext(keys ...any) func(ctx context.Context) []slog.Attr {
// 	return func(ctx context.Context) []slog.Attr {
// 		attrs := make([]slog.Attr, 0, len(keys))
// 		for _, key := range keys {
// 			attrs = append(attrs, slog.Any(key.(string), ctx.Value(key)))
// 		}
// 		return attrs
// 	}
// }

// Extract the args slice directly from context
func ExtractArgsFromContext(ctx context.Context) []slog.Attr {
	if v := ctx.Value(ctxAttrsKey{}); v != nil {
		if s, ok := v.([]slog.Attr); ok {
			return s
		}
	}
	return nil
}
