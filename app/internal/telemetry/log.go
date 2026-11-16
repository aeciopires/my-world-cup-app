package telemetry

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// NewLogger returns a structured logger writing to w as JSON (asJSON) or
// logfmt-style text, whose records carry trace_id and
// span_id whenever they're logged with a context holding a valid span
// (slog.InfoContext and friends). Those fields are what lets Grafana jump
// from a log line in Loki to its trace in Tempo/Jaeger.
func NewLogger(w io.Writer, asJSON bool) *slog.Logger {
	var h slog.Handler
	if asJSON {
		h = slog.NewJSONHandler(w, nil)
	} else {
		h = slog.NewTextHandler(w, nil)
	}
	return slog.New(traceHandler{h})
}

// traceHandler decorates another slog.Handler with the current span's IDs.
type traceHandler struct {
	slog.Handler
}

func (h traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceHandler{h.Handler.WithAttrs(attrs)}
}

func (h traceHandler) WithGroup(name string) slog.Handler {
	return traceHandler{h.Handler.WithGroup(name)}
}
