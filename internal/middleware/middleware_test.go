package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/logging"
)

func TestLoggerTraceIDPropagation(t *testing.T) {
	var ctxTraceID string
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxTraceID = logging.TraceID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-Trace-Id", "fe-trace-001")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Trace-Id"); got != "fe-trace-001" {
		t.Fatalf("expected response X-Trace-Id=fe-trace-001, got %q", got)
	}
	if ctxTraceID != "fe-trace-001" {
		t.Fatalf("expected context trace ID fe-trace-001, got %q", ctxTraceID)
	}

	// Forged header with whitespace/newlines must be rejected and replaced.
	reqBad := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	reqBad.Header.Set("X-Trace-Id", "bad trace\nline")
	recBad := httptest.NewRecorder()
	handler.ServeHTTP(recBad, reqBad)

	gotBad := recBad.Header().Get("X-Trace-Id")
	if gotBad == "" || gotBad == "bad trace\nline" {
		t.Fatalf("expected sanitized fallback trace ID, got %q", gotBad)
	}
}
