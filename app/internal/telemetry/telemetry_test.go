package telemetry

import (
	"bytes"
	"encoding/json"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func spanContext(t *testing.T) trace.SpanContext {
	t.Helper()
	traceID, err := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	if err != nil {
		t.Fatal(err)
	}
	spanID, err := trace.SpanIDFromHex("00f067aa0ba902b7")
	if err != nil {
		t.Fatal(err)
	}
	return trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled})
}

func TestNewLoggerAddsTraceIDs(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, true).With("component", "test")
	ctx := trace.ContextWithSpanContext(t.Context(), spanContext(t))

	logger.InfoContext(ctx, "hello")

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("output is not JSON: %v (%q)", err, buf.String())
	}
	if rec["trace_id"] != "4bf92f3577b34da6a3ce929d0e0e4736" || rec["span_id"] != "00f067aa0ba902b7" {
		t.Errorf("trace fields missing or wrong: %v", rec)
	}
	if rec["component"] != "test" {
		t.Errorf("WithAttrs lost through the trace handler: %v", rec)
	}
}

func TestNewLoggerWithoutSpanOmitsTraceIDs(t *testing.T) {
	var buf bytes.Buffer
	NewLogger(&buf, false).InfoContext(t.Context(), "hello")

	out := buf.String()
	if bytes.Contains(buf.Bytes(), []byte("trace_id")) {
		t.Errorf("unexpected trace_id without a span: %q", out)
	}
	if !bytes.Contains(buf.Bytes(), []byte("msg=hello")) {
		t.Errorf("expected logfmt text output, got %q", out)
	}
}

func TestEnabled(t *testing.T) {
	tests := []struct {
		name, endpoint, tracesEndpoint, disabled string
		want                                     bool
	}{
		{name: "nothing set"},
		{name: "generic endpoint", endpoint: "http://alloy:4318", want: true},
		{name: "traces endpoint", tracesEndpoint: "http://alloy:4318/v1/traces", want: true},
		{name: "explicitly disabled", endpoint: "http://alloy:4318", disabled: "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", tt.endpoint)
			t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", tt.tracesEndpoint)
			t.Setenv("OTEL_SDK_DISABLED", tt.disabled)
			if got := Enabled(); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupDisabledIsNoop(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	shutdown, err := Setup(t.Context(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := shutdown(t.Context()); err != nil {
		t.Errorf("no-op shutdown returned %v", err)
	}
}
