package xlog

import (
	"log/slog"

	"github.com/Amnesiac9/xlog"
	"github.com/labstack/echo/v4"
)

// With adds one or more key–value pairs to the logger stored in the context,
// returning a *new* echo.Context whose request.Context carries that enriched logger.
// Use it when you want subsequent log calls to automatically include those attributes.
func With(c echo.Context, attrs ...any) echo.Context {
	ctx := c.Request().Context()
	logger := xlog.LoggerFromContext(ctx).With(attrs...)
	ctx = xlog.ContextWithLogger(ctx, logger)

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
	xlog.LoggerFromContext(ctx).InfoContext(ctx, msg, args...)
}

func Info(c echo.Context, msg string, args ...any) {
	ctx := c.Request().Context()
	xlog.LoggerFromContext(ctx).InfoContext(ctx, msg, args...)
}

func Warn(c echo.Context, msg string, args ...any) {
	ctx := c.Request().Context()
	xlog.LoggerFromContext(ctx).WarnContext(ctx, msg, args...)
}

func Error(c echo.Context, msg string, err error, args ...any) {
	ctx := c.Request().Context()
	xlog.LoggerFromContext(ctx).ErrorContext(ctx, msg, append(args, slog.String("error", err.Error()))...)
}
