package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestTracerFromContext(t *testing.T) {
	t.Run("returns global tracer for invalid context", func(t *testing.T) {
		ctx := context.Background()
		tracer := TracerFromContext(ctx)
		assert.NotNil(t, tracer)
	})

	t.Run("returns span tracer for valid context", func(t *testing.T) {
		ctx := context.Background()
		tracer := otel.Tracer(TracerName)
		ctx, span := tracer.Start(ctx, "test-span")
		defer span.End()

		resultTracer := TracerFromContext(ctx)
		assert.NotNil(t, resultTracer)
	})
}

func TestSpanIDFromContext(t *testing.T) {
	t.Run("returns empty string for empty context", func(t *testing.T) {
		ctx := context.Background()
		spanID := SpanIDFromContext(ctx)
		assert.Empty(t, spanID)
	})

	t.Run("returns spanID for valid context", func(t *testing.T) {
		// Create a context with span
		tracer := sdktrace.NewTracerProvider().Tracer("test")
		ctx, span := tracer.Start(
			context.Background(),
			"foo",
			oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		)
		defer span.End()

		assert.NotEmpty(t, SpanIDFromContext(ctx))
	})
}

func TestTraceIDFromContext(t *testing.T) {
	t.Run("returns empty string for empty context", func(t *testing.T) {
		ctx := context.Background()
		traceID := TraceIDFromContext(ctx)
		assert.Empty(t, traceID)
	})

	t.Run("returns traceID for valid context", func(t *testing.T) {
		// Create a context with span
		tracer := sdktrace.NewTracerProvider().Tracer("test")
		ctx, span := tracer.Start(
			context.Background(),
			"foo",
			oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		)
		defer span.End()

		assert.NotEmpty(t, TraceIDFromContext(ctx))
	})
}
