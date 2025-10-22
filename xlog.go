package xlog

import (
	"context"
	"log/slog"
	"net/http"
)

// ---------------------------------------------------------------------
// Context helpers for request-scoped loggers
// ---------------------------------------------------------------------

// Key for the context logger to avoid collisions
type ctxLoggerKey struct{}

// Return the context with the logger added with the logger key.
func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

// Get the logger from the context with the logger key, or default logger.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

// With adds one or more key–value pairs to the logger stored in the context.
// Use it when you want subsequent log calls to automatically include those attributes.
func With(r *http.Request, attrs ...any) *http.Request {
	// ctx := With(r.Context(), attrs...)
	// return r.WithContext(ctx)

	ctx := r.Context()
	logger := LoggerFromContext(ctx).With(attrs...)
	ctx = ContextWithLogger(ctx, logger)

	// Attach the new context back to the request so later handlers/middleware see it.
	r = r.WithContext(ctx)
	return r
}

// ---------------------------------------------------------------------
// Public helpers (call these directly in handlers)
// ---------------------------------------------------------------------

func Debug(c context.Context, msg string, args ...any) {
	LoggerFromContext(c).DebugContext(c, msg, args...)
}

func Info(c context.Context, msg string, args ...any) {
	LoggerFromContext(c).InfoContext(c, msg, args...)
}

func Warn(c context.Context, msg string, args ...any) {
	LoggerFromContext(c).WarnContext(c, msg, args...)
}

func Error(c context.Context, msg string, err error, args ...any) {
	LoggerFromContext(c).ErrorContext(c, msg, append(args, slog.String("error", err.Error()))...)
}
