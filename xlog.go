package xlog

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------
// Context helpers for request-scoped loggers
// ---------------------------------------------------------------------

// Key for the context logger to avoid collisions
type ctxLoggerKey struct{}

// Return the context with the logger added with the logger key.
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
// returning a *new* echo.Context whose request.Context carries that enriched logger.
// Use it when you want subsequent log calls to automatically include those attributes.
func With(c echo.Context, attrs ...any) echo.Context {
	ctx := c.Request().Context()
	logger := FromContext(ctx).With(attrs...)
	ctx = ToContext(ctx, logger)

	// Attach the new context back to the request so later handlers/middleware see it.
	req := c.Request().WithContext(ctx)
	c.SetRequest(req)
	return c
}

// ---------------------------------------------------------------------
// Public helpers (call these directly in handlers)
// ---------------------------------------------------------------------

func Debug(c echo.Context, msg string, args ...any) {
	ctx := c.Request().Context()
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

func Info(c echo.Context, msg string, args ...any) {
	ctx := c.Request().Context()
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

func Warn(c echo.Context, msg string, args ...any) {
	ctx := c.Request().Context()
	FromContext(ctx).WarnContext(ctx, msg, args...)
}

func Error(c echo.Context, msg string, err error, args ...any) {
	ctx := c.Request().Context()
	FromContext(ctx).ErrorContext(ctx, msg, append(args, slog.String("error", err.Error()))...)
}
