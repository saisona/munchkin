package telemetry

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	DefaultSvcTracer  = otel.Tracer("munchin/svc")
	DefaultRepoTracer = otel.Tracer("munchin/repo")
)

func InitTracer(
	ctx context.Context,
	serviceName string,
	otlpEndpoint string,
) (func(context.Context) error, error) {
	// Configure OTLP exporter
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(otlpEndpoint), // e.g. "localhost:4317"
		otlptracegrpc.WithInsecure(),             // use TLS in production
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}

	// Resource describes the service
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServerPort(1337),
			semconv.ContainerImageName("app"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	// ✅ REQUIRED
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// Return shutdown function
	return tp.Shutdown, nil
}

func traceFieldsFromContext(ctx context.Context) []slog.Attr {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return nil
	}

	sc := span.SpanContext()
	return []slog.Attr{
		slog.String("trace_id", sc.TraceID().String()),
		slog.String("span_id", sc.SpanID().String()),
	}
}

type TraceHandler struct {
	slog.Handler
}

func (h TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := traceFieldsFromContext(ctx); len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, r)
}
