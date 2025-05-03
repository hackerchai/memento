package xlog

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// newEventWithTrace is a helper function to create a new Event with trace information.
func (xlog *Logger) newEventWithTrace(level Level, event *Event, ctx context.Context) *LogEvent {
	e := xlog.newEvent(level, event)

	if ctx != nil {
		if reqID, ok := ctx.Value(XLogReqIDKey(xlog.reqIDKey)).(string); ok {
			e.Str("reqID", reqID)
		}

		if span := trace.SpanFromContext(ctx); span != nil {
			spanContext := span.SpanContext()
			if spanContext.IsValid() {
				e.Str("traceID", spanContext.TraceID().String())
				e.Str("spanID", spanContext.SpanID().String())
				e.Bool("sampled", spanContext.IsSampled())
			}
		}
	}

	return e
}
