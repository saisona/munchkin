// Package telemetry provides centralized OpenTelemetry instrumentation for the
// application.
//
// It is responsible for initializing and configuring telemetry signals
// (traces and logs) and exporting them using the OTLP protocol to a local
// OpenTelemetry Collector (Grafana Alloy).
//
// The package:
//
//   - Configures distributed tracing using OpenTelemetry SDK
//   - Bridges Go's structured logging (log/slog) to OpenTelemetry Logs
//   - Ensures trace and span context propagation across application layers
//   - Integrates with HTTP middleware (Echo) for automatic span creation
//
// Telemetry data is exported via OTLP/gRPC to Grafana Alloy, which is
// responsible for routing:
//
//   - Traces → Tempo
//   - Logs   → Loki
//
// This package intentionally contains no business logic and should be
// initialized once at application startup (typically from main).
//
// Example usage:
//
//	ctx := context.Background()
//
//	tp, lp, err := telemetry.Init(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer tp.Shutdown(ctx)
//	defer lp.Shutdown(ctx)
package telemetry
