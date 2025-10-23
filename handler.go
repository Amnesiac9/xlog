package xlog

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"time"
)

type XlogHandler struct {
	option Option
	attrs  []slog.Attr
	groups []string
}

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// default: 10s
	Timeout time.Duration

	// output (default: os.Stdout)
	Writer io.Writer

	// optional: custom marshaler
	Marshaler func(v any) ([]byte, error)
	// optional: fetch attributes from context
	AttrFromContext []func(ctx context.Context) []slog.Attr

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

func (o Option) NewXlogHandler() slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelDebug
	}

	if o.Timeout == 0 {
		o.Timeout = 10 * time.Second
	}

	if o.Marshaler == nil {
		o.Marshaler = json.Marshal
	}

	if o.AttrFromContext == nil {
		o.AttrFromContext = []func(ctx context.Context) []slog.Attr{}
	}

	return &XlogHandler{
		option: o,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

var _ slog.Handler = (*XlogHandler)(nil)

func (h *XlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

func (h *XlogHandler) Handle(ctx context.Context, record slog.Record) error {
	fromContext := ContextExtractor(ctx, h.option.AttrFromContext)
	attrs := append(h.attrs, fromContext...)

	return nil
}

func (h *XlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &XlogHandler{
		option: h.option,
		attrs:  append(h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *XlogHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &XlogHandler{
		option: h.option,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

func ContextExtractor(ctx context.Context, fns []func(ctx context.Context) []slog.Attr) []slog.Attr {
	attrs := []slog.Attr{}
	for _, fn := range fns {
		attrs = append(attrs, fn(ctx)...)
	}
	return attrs
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

// func defaultRequestArgs(ctx context.Context) []slog.Attr {
// 	// Add global context attrs to log here.
// 	r := []slog.Attr{}
// 	if v := ctx.Value(CtxTenantKey); v != nil {
// 		r = append(r, slog.String(string(CtxTenantKey), fmt.Sprint(v)))
// 	}
// 	if v := ctx.Value(CtxReqIDKey); v != nil {
// 		r = append(r, slog.String(string(CtxReqIDKey), fmt.Sprint(v)))
// 	}
// 	if v := ctx.Value(CtxMethodKey); v != nil {
// 		r = append(r, slog.String(string(CtxMethodKey), fmt.Sprint(v)))
// 	}
// 	if v := ctx.Value(CtxURIPathKey); v != nil {
// 		r = append(r, slog.String(string(CtxURIPathKey), fmt.Sprint(v)))
// 	}
// 	return r
// }
