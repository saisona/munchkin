package log

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type OTelHandler struct {
	next slog.Handler
}

func NewOTelHandler(next slog.Handler) slog.Handler {
	return &OTelHandler{next: next}
}

func (h *OTelHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.next.Enabled(ctx, lvl)
}

func (h *OTelHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		sc := span.SpanContext()
		if sc.IsValid() {
			r.AddAttrs(
				slog.String("trace_id", sc.TraceID().String()),
				slog.String("span_id", sc.SpanID().String()),
			)
		}
	}
	return h.next.Handle(ctx, r)
}

func (h *OTelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OTelHandler{next: h.next.WithAttrs(attrs)}
}

func (h *OTelHandler) WithGroup(name string) slog.Handler {
	return &OTelHandler{next: h.next.WithGroup(name)}
}
