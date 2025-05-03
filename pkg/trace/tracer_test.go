package trace

//import (
//	"context"
//	"os"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestNewTracer(t *testing.T) {
//	config := &Config{
//		Name:         "test-service",
//		Provider:     ProviderStdout,
//		SamplingRate: 1.0,
//	}
//
//	tracer := NewTracer(config)
//	assert.NotNil(t, tracer)
//}
//
//func TestTracerLifecycle(t *testing.T) {
//	tests := []struct {
//		name    string
//		config  Config
//		wantErr bool
//	}{
//		{
//			name: "stdout provider",
//			config: Config{
//				Name:         "test-service",
//				Provider:     ProviderStdout,
//				SamplingRate: 1.0,
//			},
//			wantErr: false,
//		},
//		{
//			name: "file provider",
//			config: Config{
//				Name:         "test-service",
//				Provider:     ProviderFile,
//				Endpoint:     "test.log",
//				SamplingRate: 1.0,
//			},
//			wantErr: false,
//		},
//		{
//			name: "invalid provider",
//			config: Config{
//				Name:         "test-service",
//				Provider:     "invalid",
//				SamplingRate: 1.0,
//			},
//			wantErr: true,
//		},
//		{
//			name: "no provider",
//			config: Config{
//				Name: "test-service",
//			},
//			wantErr: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			tracer := NewTracer(&tt.config)
//
//			// test start tracer
//			err := tracer.Start()
//			if tt.wantErr {
//				assert.Error(t, err)
//				return
//			}
//			assert.NoError(t, err)
//			assert.NotNil(t, tracer.GetTracerProvider())
//
//			// test duplicate start tracer
//			err = tracer.Start()
//			assert.Error(t, err, "should not allow duplicate start of the same tracer")
//
//			// test stop tracer
//			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//			defer cancel()
//
//			err = tracer.Stop(ctx)
//			assert.NoError(t, err)
//
//			// clean up test file
//			if tt.config.Provider == ProviderFile {
//				os.Remove(tt.config.Endpoint)
//			}
//		})
//	}
//}
//
//func TestTracerConcurrency(t *testing.T) {
//	config := &Config{
//		Name:         "test-service",
//		Provider:     ProviderStdout,
//		SamplingRate: 1.0,
//	}
//
//	// test concurrency start and stop tracer
//	for i := 0; i < 5; i++ {
//		go func() {
//			tracer := NewTracer(config)
//			err := tracer.Start()
//			assert.NoError(t, err)
//			if err == nil {
//				ctx := context.Background()
//				err = tracer.Stop(ctx)
//				assert.NoError(t, err)
//			}
//		}()
//	}
//}
//
//func TestSetTracerName(t *testing.T) {
//	// Store original value
//	original := TracerName
//	defer func() {
//		TracerName = original
//	}()
//
//	tests := []struct {
//		name     string
//		input    string
//		expected string
//	}{
//		{
//			name:     "set new tracer name",
//			input:    "new-tracer",
//			expected: "new-tracer",
//		},
//		{
//			name:     "set empty tracer name",
//			input:    "",
//			expected: "new-tracer",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			SetTracerName(tt.input)
//			if TracerName != tt.expected {
//				t.Errorf("want %s, got %s", tt.expected, TracerName)
//			}
//		})
//	}
//}
