package xlog

import (
	"context"
	"log/slog"
)

// ---------------------------------------------------------------------
// Context helpers for request-scoped loggers
// ---------------------------------------------------------------------

// Key for the context logger to avoid collisions
type ctxLoggerKey struct{}

// Return the context with the logger added with the logger key.
// Inteded to then be stored in the request, which can then be retrieved automatically or manually.
// Helpers included in this package automatically retrieve this logger when called.
func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

// Get the logger from the context with the logger key, or default logger.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

// With adds one or more key–value pairs to the logger stored in the context,
// returning a *new* Context that carries that enriched logger.
// Use it when you want subsequent log calls to automatically include those attributes.
func With(ctx context.Context, attrs ...any) context.Context {
	logger := FromContext(ctx).With(attrs...)
	return ToContext(ctx, logger)
}

// ---------------------------------------------------------------------
// Public helpers (call these directly in handlers)
// ---------------------------------------------------------------------

func Debug(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).WarnContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, err error, args ...any) {
	FromContext(ctx).ErrorContext(ctx, msg, append(args, slog.String("error", err.Error()))...)
}
