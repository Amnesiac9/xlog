package xlog

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ctxAttrsKey struct{}

// Request-scoped slog.Logger to the context with default per-req attrs.
//
// Calling Info on this method: 319.1 ns/op	       0 B/op	       0 allocs/op
func MiddlewareAttachDefaultsLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenant := GetTenant(c)

			req := c.Request()
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)

			// Create a logger with default values
			reqLogger := logger.With(
				slog.String(string(CtxTenantKey), tenant),
				slog.String(string(CtxMethodKey), req.Method),
				slog.String(string(CtxURIPathKey), req.URL.Path),
				slog.String(string(CtxReqIDKey), reqID),
			)

			// Add the logger to context so that we can call it.
			ctx := ToContext(req.Context(), reqLogger)
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}

// Attach Default Per Request attributes to the context.
//
// Benchmark:	       503.3 ns/op	       0 B/op	       0 allocs/op
func MiddlewareAttachDefaultsCtx(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenant := GetTenant(c)

			req := c.Request()
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)

			// Allows simple slog.InfoContext calls to also return these values rather than requiring the use of xlog.Level() funcs
			// ctx := c.Request().Context()
			// ctx = context.WithValue(ctx, CtxTenantKey, tenant)
			// ctx = context.WithValue(ctx, CtxReqIDKey, reqID)
			// ctx = context.WithValue(ctx, CtxMethodKey, c.Request().Method)
			// ctx = context.WithValue(ctx, CtxURIPathKey, c.Request().URL.Path)

			// Create default attrs and store the slice:
			attrs := []slog.Attr{
				slog.String(string(CtxTenantKey), tenant),
				slog.String(string(CtxReqIDKey), reqID),
				slog.String(string(CtxMethodKey), c.Request().Method),
				slog.String(string(CtxURIPathKey), c.Request().URL.Path),
			}

			ctx := context.WithValue(req.Context(), ctxAttrsKey{}, attrs)

			//ctx := ContextWithLogger(req.Context(), reqLogger)
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}

// Attach Default Per Request attributes to the context.
//
//	Calling Info on this method: 793.2 ns/op	     336 B/op	       7 allocs/op
func MiddlewareAttachDefaultsCtxOld(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenant := GetTenant(c)

			req := c.Request()
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)

			// Allows simple slog.InfoContext calls to also return these values rather than requiring the use of xlog.Level() funcs
			ctx := req.Context()
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
// TODO: Alternative error messages for frontend?
func MiddlewareRequestLoggerSlog() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogURIPath:   true,
		LogError:     true,
		HandleError:  true,
		LogRequestID: true,
		LogMethod:    true,
		LogLatency:   true,
		LogHost:      true,
		LogRemoteIP:  true,
		LogUserAgent: true,
		LogReferer:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.Int("status", v.Status),
				slog.Int64("duration_ms", v.Latency.Milliseconds()),
			}

			logger := FromContext(c.Request().Context())

			if v.Error == nil {
				logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "REQUEST", attrs...)
			} else {
				attrs = append(attrs,
					slog.String("uri", v.URI),
					slog.String("host", v.Host),
					slog.String("remote_ip", v.RemoteIP),
					slog.String("user_agent", v.UserAgent),
					slog.String("referer", v.Referer),
					slog.String("error", v.Error.Error()),
				)
				logger.LogAttrs(c.Request().Context(), slog.LevelError, "REQUEST_ERROR", attrs...)
			}
			return nil
		},
	})
}
