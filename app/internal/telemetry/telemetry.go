// Package telemetry wires up OpenTelemetry tracing and trace-aware
// structured logging.
//
// Tracing is opt-in and configured entirely through the standard
// OpenTelemetry environment variables (OTEL_EXPORTER_OTLP_ENDPOINT,
// OTEL_SERVICE_NAME, OTEL_TRACES_SAMPLER, ...). When no OTLP endpoint is set
// the global no-op tracer stays in place, so instrumented code costs almost
// nothing and nothing is exported.
package telemetry

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ServiceName is the default OpenTelemetry service.name; OTEL_SERVICE_NAME
// overrides it.
const ServiceName = "my-world-cup-app"

// ShutdownFunc flushes buffered spans and releases exporter resources.
type ShutdownFunc func(context.Context) error

// Enabled reports whether tracing is configured: an OTLP endpoint is set and
// the SDK isn't explicitly disabled via OTEL_SDK_DISABLED.
func Enabled() bool {
	if disabled, _ := strconv.ParseBool(os.Getenv("OTEL_SDK_DISABLED")); disabled {
		return false
	}
	return cmp.Or(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"), os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) != ""
}

// Setup installs the W3C trace-context propagator and, when Enabled, a
// global TracerProvider exporting spans over OTLP/HTTP. The returned
// ShutdownFunc must be called before the process exits so buffered spans
// are flushed; it is a no-op when tracing is disabled.
func Setup(ctx context.Context, version string) (ShutdownFunc, error) {
	// Propagation is installed even with tracing off, so an incoming
	// traceparent header still lands in logs and is forwarded upstream.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	if !Enabled() {
		return func(context.Context) error { return nil }, nil
	}

	// The exporter reads the endpoint, headers, timeout, etc. from the
	// OTEL_EXPORTER_OTLP_* environment variables itself.
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	// Later options take precedence, so OTEL_SERVICE_NAME and
	// OTEL_RESOURCE_ATTRIBUTES (WithFromEnv) override the defaults.
	res, err := resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			attribute.String("service.name", ServiceName),
			attribute.String("service.version", version),
		),
		resource.WithFromEnv(),
	)
	if err != nil && !errors.Is(err, resource.ErrPartialResource) {
		return nil, fmt.Errorf("build OpenTelemetry resource: %w", err)
	}

	// No explicit sampler: the SDK honors OTEL_TRACES_SAMPLER(_ARG) and
	// defaults to parentbased_always_on.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
