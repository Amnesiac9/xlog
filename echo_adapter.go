package xlog

import (
	"log/slog"

	"github.com/labstack/echo/v4"
)

// With adds one or more key–value pairs to the logger stored in the context,
// returning a *new* echo.Context whose request.Context carries that enriched logger.
// Use it when you want subsequent log calls to automatically include those attributes.
func WithC(c echo.Context, attrs ...any) echo.Context {
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

func DebugC(c echo.Context, msg string, args ...any) {
	ctx := c.Request().Context()
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

func InfoC(c echo.Context, msg string, args ...any) {
	ctx := c.Request().Context()
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

func WarnC(c echo.Context, msg string, args ...any) {
	ctx := c.Request().Context()
	FromContext(ctx).WarnContext(ctx, msg, args...)
}

func ErrorC(c echo.Context, msg string, err error, args ...any) {
	ctx := c.Request().Context()
	FromContext(ctx).ErrorContext(ctx, msg, append(args, slog.String("error", err.Error()))...)
}
