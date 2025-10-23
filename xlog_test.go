package xlog

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func Test_LoggingLevels(t *testing.T) {
	InitLogger()

	// Build a context with your per-request values
	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxTenantKey, "lecole-no41")
	ctx = context.WithValue(ctx, CtxReqIDKey, "req-123")
	ctx = context.WithValue(ctx, CtxMethodKey, "GET")
	ctx = context.WithValue(ctx, CtxURIPathKey, "/api/cards/lookup")

	slog.InfoContext(ctx, "Test Info", slog.String("extra", "yes"))
	slog.WarnContext(ctx, "something odd", slog.Int("code", 42))
}

func InitLogger() *slog.Logger {
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				// Convert the level value to a string and then to lowercase
				level := a.Value.Any().(slog.Level)
				a.Value = slog.StringValue(strings.ToLower(level.String()))
			}
			return a
		},
	})

	ctxHandler := NewXlogHandler(stdoutHandler, DefaultPerRequestArgs)

	logger := slog.New(ctxHandler)
	slog.SetDefault(logger)
	return logger
}
