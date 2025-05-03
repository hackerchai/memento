package xlog

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger()
	assert.NotNil(t, logger)
	assert.Equal(t, DefaultLogLevel, logger.GetLevel())
	assert.Equal(t, DefaultCallerDepth, logger.GetCallerDepth())
}

func TestLogger_SetLevel(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetLevel(Ldebug)
	assert.Equal(t, Ldebug, logger.GetLevel())
}

func TestLogger_SetCaller(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetCaller(3)
	assert.Equal(t, 3, logger.GetCallerDepth())
}

func TestLogger_SetTimeField(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetTimeField("custom_time")
	assert.Equal(t, "custom_time", logger.GetTimeField())
}

func TestLogger_SetTimeFormat(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetTimeFormat(TimeFormatUnixMs)
	assert.Equal(t, TimeFormatUnixMs, logger.GetTimeFormat())
}

func TestLogger_SetTimeLocation(t *testing.T) {
	logger := NewDefaultLogger()
	loc, _ := time.LoadLocation("America/New_York")
	logger.SetTimeLocation(loc)
	assert.Equal(t, loc, logger.GetTimeLocation())
}

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger()
	logger.SetWriter(&buf)

	// Create new logger and add fields
	newLogger := logger.With().
		Str("key", "value").
		Logger()

	newLogger.InfoX().Msg("test message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "test message", logEntry["message"])
	assert.Equal(t, "value", logEntry["key"])
}

func TestLogger_Build(t *testing.T) {
	logger := NewDefaultLogger()
	built := logger.Build()
	assert.Equal(t, logger, built)
}

// Note: FatalX, PanicX, FatalMsg, and PanicMsg are not tested here
// as they would terminate the program. Consider mocking these functions
// for thorough testing if necessary.

func TestLogger_LogMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger()
	logger.SetWriter(&buf)
	logger.SetLevel(Ltrace)

	tests := []struct {
		name     string
		logFunc  func(string)
		logLevel string
	}{
		{"TraceMsg", logger.TraceMsg, "trace"},
		{"DebugMsg", logger.DebugMsg, "debug"},
		{"InfoMsg", logger.InfoMsg, "info"},
		{"WarnMsg", logger.WarnMsg, "warn"},
		{"ErrorMsg", logger.ErrorMsg, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc("test message")
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)
			assert.Equal(t, tt.logLevel, logEntry["level"])
			assert.Equal(t, "test message", logEntry["message"])
		})
	}
}

func TestLogger_LogMethodsX(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger()
	logger.SetWriter(&buf)
	logger.SetLevel(Ltrace)

	tests := []struct {
		name     string
		logFunc  func(ctx ...context.Context) *LogEvent
		logLevel string
	}{
		{"TraceX", logger.TraceX, "trace"},
		{"DebugX", logger.DebugX, "debug"},
		{"InfoX", logger.InfoX, "info"},
		{"WarnX", logger.WarnX, "warn"},
		{"ErrorX", logger.ErrorX, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc().Msg("test message")
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)
			assert.Equal(t, tt.logLevel, logEntry["level"])
			assert.Equal(t, "test message", logEntry["message"])
		})
	}
}

func TestLogger_InfoXWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger()
	logger.SetWriter(&buf)

	logger.InfoX().
		Str("strField", "testString").
		Int("intField", 42).
		Bool("boolField", true).
		Float64("floatField", 3.14).
		Msg("multiple fields test")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "multiple fields test", logEntry["message"])
	assert.Equal(t, "testString", logEntry["strField"])
	assert.Equal(t, float64(42), logEntry["intField"])
	assert.Equal(t, true, logEntry["boolField"])
	assert.Equal(t, 3.14, logEntry["floatField"])
}

// Mock Hook implementation
type testHook struct {
	ran     bool
	level   Level
	message string
}

func (h *testHook) Run(event *Event, level Level, msg string) {
	h.ran = true
	h.level = level
	h.message = msg
	event.Str("hook_added", "true")
}

func TestLogger_InfoMsg_WithHook(t *testing.T) {
	// Prepare test data
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, Linfo, TimeFormatUnix, nil)

	// Create and add test Hook
	hook := &testHook{}
	logger.AddHook(hook)

	// Execute the method being tested
	testMsg := "test info message"
	logger.InfoMsg(testMsg)

	// Verify if the Hook was executed correctly
	assert.True(t, hook.ran, "Hook should have been executed")
	assert.Equal(t, Linfo, hook.level, "Hook should receive the correct log level")
	assert.Equal(t, testMsg, hook.message, "Hook should receive the correct log message")

	// Verify log output
	logOutput := buf.String()
	assert.Contains(t, logOutput, testMsg, "Log output should contain the original message")
	assert.Contains(t, logOutput, "hook_added", "Log output should contain the field added by the Hook")
	assert.Contains(t, logOutput, "true", "Log output should contain the value added by the Hook")
}

func TestLogger_InfoX_WithHook(t *testing.T) {
	// Create a test Logger
	var buf bytes.Buffer
	testLogger := NewLogger(&buf, Linfo, TimeFormatUnix, nil)

	// Create a test Hook
	testHook := &testHook{}

	// Add Hook to Logger
	testLogger.AddHook(testHook)

	// Call InfoX and write a message
	testMsg := "test info message"
	testLogger.InfoX().Msg(testMsg)

	// Check output
	logOutput := buf.String()
	assert.Contains(t, logOutput, testMsg, "Log output should contain the original message")
	assert.Contains(t, logOutput, "hook_added", "Log output should contain the field added by the Hook")
	assert.Contains(t, logOutput, "true", "Log output should contain the value added by the Hook")
}

func TestLogger_ReqIDFromContext(t *testing.T) {
	// Create a new Logger instance
	logger := NewDefaultLogger()

	// Test cases
	testCases := []struct {
		name     string
		reqID    string
		expected string
	}{
		{
			name:     "Valid request ID",
			reqID:    "test-req-id-123",
			expected: "test-req-id-123",
		},
		{
			name:     "Empty request ID",
			reqID:    "",
			expected: "",
		},
		{
			name:     "Request ID not present",
			reqID:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.reqID != "" {
				ctx = logger.SetReqIDToContext(ctx, tc.reqID)
			}

			result := logger.ReqIDFromContext(ctx)

			if result != tc.expected {
				t.Errorf("Expected %s, but got %s", tc.expected, result)
			}
		})
	}
}

func TestLogger_SetReqIDToContext(t *testing.T) {
	// Create a new Logger instance
	logger := NewDefaultLogger()

	// Test cases
	testCases := []struct {
		name  string
		reqID string
	}{
		{
			name:  "Set valid request ID",
			reqID: "test-req-id-456",
		},
		{
			name:  "Set empty request ID",
			reqID: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx := logger.SetReqIDToContext(ctx, tc.reqID)

			// Verify that the new context contains the correct request ID
			result := logger.ReqIDFromContext(newCtx)

			if result != tc.reqID {
				t.Errorf("Expected %s, but got %s", tc.reqID, result)
			}

			// Verify that the original context was not modified
			originalResult := logger.ReqIDFromContext(ctx)
			if originalResult != "" {
				t.Errorf("Original context should not contain request ID, but got %s", originalResult)
			}
		})
	}
}
