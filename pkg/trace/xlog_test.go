package trace

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hackerchai/memento/pkg/xlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func TestTracerIntegrationWithXLog(t *testing.T) {
	// Setup tracer with stdout provider
	// Set sampling rate to 1.0 to ensure all traces are sampled
	cfg := NewTracerConfig().
		WithName("test-tracer").
		WithProvider(ProviderStdout).
		WithSamplingRate(1.0)

	// Initialize tracer
	tracer := NewTracer(cfg)
	err := tracer.Start()
	require.NoError(t, err)
	defer func() {
		err := tracer.Stop(context.Background())
		require.NoError(t, err)
	}()

	t.Run("with trace information", func(t *testing.T) {
		// Create buffer for log output
		var buf bytes.Buffer

		// Create logger with buffer
		logger := xlog.NewLogger(&buf, xlog.Ldebug, "", time.UTC)

		// Get tracer
		tracer := otel.GetTracerProvider().Tracer("test-tracer")

		// Create context with span
		ctx, span := tracer.Start(context.Background(), "test-operation")
		defer span.End()

		// Set request ID to context
		reqID := "test-req-123"
		ctx = logger.SetReqIDToContext(ctx, reqID)

		// Log message
		logger.InfoX(ctx).Msg("test traced message")

		// Parse log output
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		// Verify log content
		assert.Equal(t, "test traced message", logEntry["message"])
		assert.Equal(t, reqID, logEntry["reqID"])
		assert.NotEmpty(t, logEntry["traceID"])
		assert.NotEmpty(t, logEntry["spanID"])
		assert.Contains(t, logEntry, "sampled")
		assert.Equal(t, "info", logEntry["level"])

		// Verify trace context
		assert.Equal(t, TraceIDFromContext(ctx), logEntry["traceID"])
		assert.Equal(t, SpanIDFromContext(ctx), logEntry["spanID"])
	})
}
