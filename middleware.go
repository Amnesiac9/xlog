package xlog

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
)

// // Request-scoped slog.Logger to the context with attrs.
func MiddlewareAttachLoggerDefaults(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenant := GetTenant(c)

			req := c.Request()
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)

			// reqLogger := logger.With(
			// 	slog.String("tenant", tenant),
			// 	slog.String("method", req.Method),
			// 	slog.String("uri", req.URL.Path),
			// 	slog.String("request_id", reqID),
			// )

			// Allows simple slog.InfoContext calls to also return these values rather than requiring the use of xlog.Level() funcs
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, CtxTenantKey, tenant)
			ctx = context.WithValue(ctx, CtxReqIDKey, reqID)
			ctx = context.WithValue(ctx, CtxMethodKey, c.Request().Method)
			ctx = context.WithValue(ctx, CtxURIPathKey, c.Request().URL.Path)

			//ctx := ContextWithLogger(req.Context(), reqLogger)
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}
