package xlog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func Test_LoggingLevels(t *testing.T) {
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				// Convert the level value to a string and then to lowercase
				level := a.Value.Any().(slog.Level)
				a.Value = slog.StringValue(strings.ToLower(level.String()))
			}
			return a
		},
	})

	ctxHandler := NewXlogHandler(stdoutHandler, DefaultPerRequestArgs)

	logger := slog.New(ctxHandler)
	slog.SetDefault(logger)

	// Build a context with your per-request values
	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxTenantKey, "test-tenant")
	ctx = context.WithValue(ctx, CtxReqIDKey, "req-123")
	ctx = context.WithValue(ctx, CtxMethodKey, "GET")
	ctx = context.WithValue(ctx, CtxURIPathKey, "/api/cards/lookup")

	slog.InfoContext(ctx, "Test Info", slog.String("extra", "yes"))
	slog.WarnContext(ctx, "something odd", slog.Int("code", 42))
}

// newTestLogger wires XlogHandler -> JSONHandler -> buffer so we can assert on output.
func newTestLogger(w *bytes.Buffer) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		// If you use ReplaceAttr/AddSource in prod, you can include them here too.
	})
	return slog.New(NewXlogHandler(h, DefaultPerRequestArgs))
}

func Test_XlogHandler_DefaultsAppear(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	// Build a context with your per-request values
	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxTenantKey, "test-tenant")
	ctx = context.WithValue(ctx, CtxReqIDKey, "req-123")
	ctx = context.WithValue(ctx, CtxMethodKey, "GET")
	ctx = context.WithValue(ctx, CtxURIPathKey, "/api/cards/lookup")

	// Log a couple entries at different levels
	logger.InfoContext(ctx, "hello world", slog.String("extra", "yes"))
	logger.WarnContext(ctx, "something odd", slog.Int("code", 42))

	// Parse the two JSON lines and assert expected keys
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d\n%s", len(lines), buf.String())
	}

	type rec map[string]any

	var r1, r2 rec
	if err := json.Unmarshal(lines[0], &r1); err != nil {
		t.Fatalf("unmarshal line1: %v\n%s", err, lines[0])
	}
	if err := json.Unmarshal(lines[1], &r2); err != nil {
		t.Fatalf("unmarshal line2: %v\n%s", err, lines[1])
	}

	// Helper to assert a string field
	assertStr := func(m rec, key, want string) {
		got, _ := m[key].(string)
		if got != want {
			t.Errorf("want %s=%q, got %q (record: %+v)", key, want, got, m)
		}
	}

	assertStr(r1, "tenant", "test-tenant")
	assertStr(r1, "request_id", "req-123")
	assertStr(r1, "method", "GET")
	assertStr(r1, "path", "/api/cards/lookup")
	assertStr(r1, "msg", "hello world")
	assertStr(r1, "extra", "yes")

	assertStr(r2, "tenant", "test-tenant")
	assertStr(r2, "request_id", "req-123")
	assertStr(r2, "msg", "something odd")

	// Optional: ensure level is present
	if _, ok := r1["level"]; !ok {
		t.Error("expected level in record 1")
	}
	if _, ok := r2["level"]; !ok {
		t.Error("expected level in record 2")
	}
}

func Test_XlogHandler_WithAttrsIsPreserved(t *testing.T) {
	var buf bytes.Buffer

	base := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	// Add a static attr via WithAttrs and ensure XlogHandler preserves it
	h := NewXlogHandler(base.WithAttrs([]slog.Attr{slog.String("app", "marsbytes-api")}), DefaultPerRequestArgs)
	logger := slog.New(h)

	ctx := context.WithValue(context.Background(), CtxTenantKey, "egyptian-thread-company")
	logger.InfoContext(ctx, "ping")

	var rec map[string]any
	if err := json.Unmarshal(bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))[0], &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if rec["app"] != "marsbytes-api" {
		t.Errorf("expected app=marsbytes-api, got %v", rec["app"])
	}
	if rec["tenant"] != "egyptian-thread-company" {
		t.Errorf("expected tenant=egyptian-thread-company, got %v", rec["tenant"])
	}
}
