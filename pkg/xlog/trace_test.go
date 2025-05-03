package xlog

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestLogger_InfoX_Trace(t *testing.T) {
	testCases := []struct {
		name           string
		trace          bool
		provideCtx     bool
		expectSpanInfo bool
	}{
		{
			name:           "trace enabled and context provided",
			trace:          true,
			provideCtx:     true,
			expectSpanInfo: true,
		},
		{
			name:           "trace enabled but no context provided",
			trace:          true,
			provideCtx:     false,
			expectSpanInfo: false,
		},
		{
			name:           "trace disabled but context provided",
			trace:          false,
			provideCtx:     true,
			expectSpanInfo: false,
		},
		{
			name:           "trace disabled and no context provided",
			trace:          false,
			provideCtx:     false,
			expectSpanInfo: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var buf bytes.Buffer

			// Create a Logger
			logger := NewLogger(&buf, Linfo, TimeFormatUnix, nil).SetTrace(tc.trace)

			// Create a context with a span
			ctx, span := trace.NewTracerProvider().Tracer("test").Start(context.Background(), "test-span")
			defer span.End()

			// Call InfoX
			var event *LogEvent
			if tc.provideCtx {
				event = logger.InfoX(ctx)
			} else {
				event = logger.InfoX()
			}

			// Write a message to trigger log output
			event.Msg("test message")

			// Verify results
			output := buf.String()
			hasTraceID := strings.Contains(output, "traceID")
			hasSpanID := strings.Contains(output, "spanID")

			if tc.expectSpanInfo {
				if !hasTraceID || !hasSpanID {
					t.Errorf("Expected output to contain traceID and spanID, but actual output was: %s", output)
				}
			} else {
				if hasTraceID || hasSpanID {
					t.Errorf("Did not expect output to contain traceID and spanID, but actual output was: %s", output)
				}
			}
		})
	}
}

func TestLoggerWithReqID(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a new logger with output writing to the buffer
	logger := NewLogger(&buf, Linfo, TimeFormatUnix, nil)

	// Create a context with reqID
	ctx := logger.SetReqIDToContext(context.Background(), "test-req-id")

	// Log a message using InfoX method with context
	logger.InfoX(ctx).Msg("Test message with reqID")

	// Parse the log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err, "Failed to unmarshal log entry")

	// Verify that reqID exists and has the correct value
	reqID, exists := logEntry["reqID"]
	assert.True(t, exists, "reqID field should exist in log entry")
	assert.Equal(t, "test-req-id", reqID, "reqID value should match")

	// Verify that the log message is correct
	msg, exists := logEntry["message"]
	assert.True(t, exists, "message field should exist in log entry")
	assert.Equal(t, "Test message with reqID", msg, "Log message should match")
}
