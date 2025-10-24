package xlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
)

// Benchmark 1: store a *logger* in context (pre-bound with attrs via With(...))
func BenchmarkLoggerInContext(b *testing.B) {
	// Base handler that encodes but writes to /dev/null to keep it realistic
	base := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	baseLogger := slog.New(base)

	// Simulate a request's context & attributes
	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxTenantKey, "test-tenant")
	ctx = context.WithValue(ctx, CtxReqIDKey, "req-123")
	ctx = context.WithValue(ctx, CtxMethodKey, "GET")
	ctx = context.WithValue(ctx, CtxURIPathKey, "/api/orders")

	// Pre-bind (once) on the per-request logger, then store logger in
	reqLogger := baseLogger.With(
		slog.String("tenant", "test-tenant"),
		slog.String("method", "GET"),
		slog.String("uri", "/api/orders"),
		slog.String("request_id", "req-123"),
	)

	ctx = ToContext(ctx, reqLogger)

	b.ReportAllocs()

	for b.Loop() {
		// Typical call pattern inside handlers:
		FromContext(ctx).InfoContext(ctx, "processing request")
	}
}

// Benchmark 2: store *values* in context, inject via XlogHandler on each call
func BenchmarkAttrsInContext_XlogHandler(b *testing.B) {
	// Wrapped handler: XlogHandler adds attrs each Handle from ctx
	inner := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	wrapped := NewHandler(inner, DefaultPerRequestArgs)
	logger := slog.New(wrapped)

	// Simulate a request's context & attributes
	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxTenantKey, "test-tenant")
	ctx = context.WithValue(ctx, CtxReqIDKey, "req-123")
	ctx = context.WithValue(ctx, CtxMethodKey, "GET")
	ctx = context.WithValue(ctx, CtxURIPathKey, "/api/orders")

	b.ReportAllocs()

	for b.Loop() {
		// Standard slog usage; XlogHandler injects attrs each call
		logger.InfoContext(ctx, "processing request")
	}
}

// Benchmark 3: store args slice in context, inject via XlogHandler on each call
func BenchmarkAttrsSliceInContext_XlogHandler(b *testing.B) {
	// Wrapped handler: XlogHandler adds attrs each Handle from ctx
	inner := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	wrapped := NewHandler(inner, ExtractArgsFromContext)
	logger := slog.New(wrapped)

	// Simulate a request's context & attributes

	// ctx = context.WithValue(ctx, CtxTenantKey, "test-tenant")
	// ctx = context.WithValue(ctx, CtxReqIDKey, "req-123")
	// ctx = context.WithValue(ctx, CtxMethodKey, "GET")
	// ctx = context.WithValue(ctx, CtxURIPathKey, "/api/orders")

	// Create default attrs and store in context:
	attrs := []slog.Attr{
		slog.String(string(CtxTenantKey), "test-tenant"),
		slog.String(string(CtxReqIDKey), "req-123"),
		slog.String(string(CtxMethodKey), "GET"),
		slog.String(string(CtxURIPathKey), "/api/orders"),
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxAttrsKey{}, attrs)

	b.ReportAllocs()

	for b.Loop() {
		// Standard slog usage; XlogHandler injects attrs each call
		logger.InfoContext(ctx, "processing request")
	}
}

// Optional: sub-benchmark varying number of attrs (useful if you want to see
// how cost scales with context-attr count).
// func BenchmarkAttrsInContext_VaryCount(b *testing.B) {
// 	makeCtx := func(n int) context.Context {
// 		ctx := context.Background()
// 		// Always include the 4 you care about, then add more dummy attrs
// 		ctx = context.WithValue(ctx, CtxTenantKey, "test-tenant")
// 		ctx = context.WithValue(ctx, CtxReqIDKey, "req-123")
// 		ctx = context.WithValue(ctx, CtxMethodKey, "GET")
// 		ctx = context.WithValue(ctx, CtxURIPathKey, "/api/orders")
// 		// Add (n-4) extra attrs to simulate growth
// 		for i := 0; i < n-4; i++ {
// 			ctx = context.WithValue(ctx, ctxKey("k"+string(rune(i))), "v")
// 		}
// 		return ctx
// 	}

// 	inner := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
// 		Level: slog.LevelInfo,
// 	})
// 	wrapped := NewHandler(inner, DefaultPerRequestArgs)
// 	logger := slog.New(wrapped)

// 	for _, n := range []int{4, 8, 16, 32} {
// 		b.Run("attrs="+itoa(n), func(b *testing.B) {
// 			ctx := makeCtx(n)
// 			b.ReportAllocs()
// 			b.ResetTimer()
// 			for b.Loop() {
// 				logger.InfoContext(ctx, "processing request")
// 			}
// 		})
// 	}
// }

// Small, allocation-free itoa for sub-benchmark names
func itoa(i int) string {
	// fast path for small sets we use above
	switch i {
	case 4:
		return "4"
	case 8:
		return "8"
	case 16:
		return "16"
	case 32:
		return "32"
	default:
		return fmtInt(i)
	}
}

func fmtInt(i int) string {
	// not performance critical—only for naming the sub-benchmarks
	return fmt.Sprintf("%d", i)
}
