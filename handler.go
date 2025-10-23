package xlog

import (
	"context"
	"fmt"
	"log/slog"
)

type XlogHandler struct {
	handler slog.Handler

	// function to add specific attributes/fields from a given context
	attrFromContext []func(context.Context) []slog.Attr
}

func NewXlogHandler(handler slog.Handler, attrFromContextFuncs ...func(context.Context) []slog.Attr) *XlogHandler {
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

func ExtractFromContext(keys ...any) func(ctx context.Context) []slog.Attr {
	return func(ctx context.Context) []slog.Attr {
		attrs := make([]slog.Attr, 0, len(keys))
		for _, key := range keys {
			attrs = append(attrs, slog.Any(key.(string), ctx.Value(key)))
		}
		return attrs
	}
}

//////////

// ctxKey is unexported to avoid collisions with other packages.
type ctxKey string

const (
	CtxTenantKey  ctxKey = "tenant"
	CtxReqIDKey   ctxKey = "request_id"
	CtxUserKey    ctxKey = "user" // optional: add whatever you like
	CtxMethodKey  ctxKey = "method"
	CtxURIPathKey ctxKey = "path"
	CtxURIKey     ctxKey = "uri"
)

func DefaultPerRequestArgs(ctx context.Context) []slog.Attr {
	// Add global context attrs to log here.
	r := []slog.Attr{}
	if v := ctx.Value(CtxTenantKey); v != nil {
		r = append(r, slog.String(string(CtxTenantKey), fmt.Sprint(v)))
	}
	if v := ctx.Value(CtxReqIDKey); v != nil {
		r = append(r, slog.String(string(CtxReqIDKey), fmt.Sprint(v)))
	}
	if v := ctx.Value(CtxMethodKey); v != nil {
		r = append(r, slog.String(string(CtxMethodKey), fmt.Sprint(v)))
	}
	if v := ctx.Value(CtxURIPathKey); v != nil {
		r = append(r, slog.String(string(CtxURIPathKey), fmt.Sprint(v)))
	}
	return r
}

// ContextHandler is a slog.Handler middleware that adds attributes from context.
// type ContextHandler struct {
// 	slog.Handler
// }

// Handle implements slog.Handler.
// func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
// 	// Add attributes from context
// 	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok {
// 		r.Add(slog.String(string(SessionIDKey), sessionID))
// 	}
// 	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
// 		r.Add(slog.String(string(RequestIDKey), requestID))
// 	}

// 	// Now pass to the next handler in the chain
// 	return h.Handler.Handle(ctx, r)
// }
