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

func Debug(c echo.Context, msg string, args []slog.Attr) {
	ctx := c.Request().Context()
	FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, msg, args...)
}

func Info(c echo.Context, msg string, args []slog.Attr) {
	ctx := c.Request().Context()
	FromContext(ctx).LogAttrs(ctx, slog.LevelInfo, msg, args...)
}

func Warn(c echo.Context, msg string, args []slog.Attr) {
	ctx := c.Request().Context()
	FromContext(ctx).LogAttrs(ctx, slog.LevelWarn, msg, args...)
}

func Error(c echo.Context, msg string, err error, args []slog.Attr) {
	ctx := c.Request().Context()
	FromContext(ctx).LogAttrs(ctx, slog.LevelError, msg, append(args, slog.String("error", err.Error()))...)
}
