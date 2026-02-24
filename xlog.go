package xlog

import (
	"context"
	"log/slog"
)

// Key for With-managed deduped attrs stored separately from the logger.
type ctxWithAttrsKey struct{}

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
// Any attrs added via With are applied on top of the base logger.
func FromContext(ctx context.Context) *slog.Logger {
	l := slog.Default()
	if stored, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger); ok && stored != nil {
		l = stored
	}
	if attrs := withAttrsFromContext(ctx); len(attrs) > 0 {
		args := make([]any, len(attrs))
		for i, a := range attrs {
			args[i] = a
		}
		l = l.With(args...)
	}
	return l
}

// With adds or updates key–value pairs on the logger carried in the context.
// If a key already exists from a previous With call, its value is replaced
// so that duplicate keys are never emitted.
func With(ctx context.Context, args ...any) context.Context {
	existing := withAttrsFromContext(ctx)
	incoming := argsToAttrs(args)
	merged := mergeAttrs(existing, incoming)
	return context.WithValue(ctx, ctxWithAttrsKey{}, merged)
}

// withAttrsFromContext returns the deduped attrs managed by With.
func withAttrsFromContext(ctx context.Context) []slog.Attr {
	if v, ok := ctx.Value(ctxWithAttrsKey{}).([]slog.Attr); ok {
		return v
	}
	return nil
}

// argsToAttrs converts slog-style args (key, value, key, value, …) into a
// slice of slog.Attr, following the same conventions as slog.Logger.With.
func argsToAttrs(args []any) []slog.Attr {
	var attrs []slog.Attr
	for len(args) > 0 {
		switch v := args[0].(type) {
		case slog.Attr:
			attrs = append(attrs, v)
			args = args[1:]
		case string:
			if len(args) < 2 {
				attrs = append(attrs, slog.String(v, "!MISSING"))
				args = args[1:]
			} else {
				attrs = append(attrs, slog.Any(v, args[1]))
				args = args[2:]
			}
		default:
			attrs = append(attrs, slog.Any("!BADKEY", v))
			args = args[1:]
		}
	}
	return attrs
}

// mergeAttrs merges incoming attrs into existing, replacing by key.
func mergeAttrs(existing, incoming []slog.Attr) []slog.Attr {
	result := make([]slog.Attr, len(existing))
	copy(result, existing)
	for _, na := range incoming {
		replaced := false
		for i, ea := range result {
			if ea.Key == na.Key {
				result[i] = na
				replaced = true
				break
			}
		}
		if !replaced {
			result = append(result, na)
		}
	}
	return result
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
