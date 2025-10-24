package xlog

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

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

// Example on how to pull individual args from context
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

func DefaultReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		// Convert the level value to a string and then to lowercase
		level := a.Value.Any().(slog.Level)
		a.Value = slog.StringValue(strings.ToLower(level.String()))
	}
	return a
}
