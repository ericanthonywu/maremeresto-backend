// Package logging carries a per-request trace ID through context.Context so
// every log line written for a request, from the HTTP access log down to the
// SQL tracer, can be correlated with one grep.
package logging

import (
	"context"
	"log/slog"
)

type traceKey struct{}

// WithTraceID returns a context carrying the given trace ID.
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceKey{}, id)
}

// TraceID returns the trace ID stored in ctx, or "" when there is none.
func TraceID(ctx context.Context) string {
	id, _ := ctx.Value(traceKey{}).(string)
	return id
}

// Handler wraps another slog.Handler and adds a trace_id attribute to every
// record logged with a context that carries one. Use the slog.*Context
// functions (slog.InfoContext, slog.ErrorContext, ...) to benefit from it.
type Handler struct {
	slog.Handler
}

func NewHandler(inner slog.Handler) *Handler { return &Handler{Handler: inner} }

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	if id := TraceID(ctx); id != "" {
		r.AddAttrs(slog.String("trace_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{Handler: h.Handler.WithGroup(name)}
}
