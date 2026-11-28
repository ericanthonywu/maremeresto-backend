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

	// Client-supplied X-Trace-Id must be ignored so callers cannot spoof trace IDs.
	reqSpoof := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	reqSpoof.Header.Set("X-Trace-Id", "fe-spoof-001")
	recSpoof := httptest.NewRecorder()
	handler.ServeHTTP(recSpoof, reqSpoof)

	gotSpoof := recSpoof.Header().Get("X-Trace-Id")
	if gotSpoof == "" || gotSpoof == "fe-spoof-001" {
		t.Fatalf("expected server-generated trace ID ignoring client X-Trace-Id, got %q", gotSpoof)
	}
	if ctxTraceID != gotSpoof {
		t.Fatalf("expected context trace ID %q, got %q", gotSpoof, ctxTraceID)
	}

	// Reverse-proxy injected X-Request-Id (e.g. Nginx $request_id) is preserved.
	reqProxy := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	reqProxy.Header.Set("X-Request-Id", "nginx-req-001")
	reqProxy.Header.Set("X-Trace-Id", "fe-spoof-002")
	recProxy := httptest.NewRecorder()
	handler.ServeHTTP(recProxy, reqProxy)

	if got := recProxy.Header().Get("X-Trace-Id"); got != "nginx-req-001" {
		t.Fatalf("expected response X-Trace-Id=nginx-req-001, got %q", got)
	}
	if ctxTraceID != "nginx-req-001" {
		t.Fatalf("expected context trace ID nginx-req-001, got %q", ctxTraceID)
	}

	// Forged X-Request-Id with whitespace/newlines must be rejected and replaced.
	reqBad := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	reqBad.Header.Set("X-Request-Id", "bad trace\nline")
	recBad := httptest.NewRecorder()
	handler.ServeHTTP(recBad, reqBad)

	gotBad := recBad.Header().Get("X-Trace-Id")
	if gotBad == "" || gotBad == "bad trace\nline" {
		t.Fatalf("expected sanitized fallback trace ID, got %q", gotBad)
	}
}
