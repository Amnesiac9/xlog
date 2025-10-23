package xlog

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

// Per request final log for echo
func MiddlewareRequestLoggerSlog(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogError:     true,
		HandleError:  true,
		LogRequestID: true,
		LogMethod:    true,
		LogLatency:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			//tenant := xlog.GetTenant(c)

			attrs := []slog.Attr{
				slog.Int("status", v.Status),
				slog.Int64("duration_ms", v.Latency.Milliseconds()),
			}

			// // Add a logger to context to add these default attrs to each log message
			// reqLogger := logger.With(
			// 	slog.String("tenant", tenant),
			// 	slog.String("method", v.Method),
			// 	slog.String("uri", v.URI),
			// 	slog.String("request_id", v.RequestID),
			// )
			// ctx := context.WithValue(c.Request().Context(), ctxLoggerKey{}, reqLogger)
			// c.SetRequest(c.Request().WithContext(ctx))

			if v.Error == nil {
				logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "REQUEST", attrs...)
			} else {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
				logger.LogAttrs(c.Request().Context(), slog.LevelError, "REQUEST_ERROR", attrs...)
			}
			return nil
		},
	})
}
