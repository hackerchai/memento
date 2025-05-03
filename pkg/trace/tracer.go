package trace

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger" //nolint:staticcheck
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// TracerName is the name of the tracer.
var (
	TracerName = "tracer"
)

const (
	ProviderOtlpGrpc = "otlp_grpc"
	ProviderOtlpHttp = "otlp_http"
	ProviderJaeger   = "jaeger"
	ProviderZipkin   = "zipkin"
	ProviderFile     = "file"
	ProviderStdout   = "stdout"
)

type Tracer interface {
	Start() error
	Stop(ctx context.Context) error
	GetTracerProvider() *sdktrace.TracerProvider
}

type tracer struct {
	tp     *sdktrace.TracerProvider
	config *Config
	mu     sync.RWMutex
}

var (
	// activeTracers keeps track of running tracers avoid duplicate initialization
	activeTracers = make(map[string]struct{})
	tracerMu      sync.RWMutex
)

// SetTracerName sets the name of the tracer.
func SetTracerName(name string) {
	if name == "" {
		return
	}

	TracerName = name
}

// NewTracer creates a new tracer.
func NewTracer(c *Config) Tracer {
	return &tracer{config: c}
}

// Start starts the tracer.
func (t *tracer) Start() error {
	// Create unique key for the tracer using provider and endpoint
	key := fmt.Sprintf("%s:%s", t.config.Provider, t.config.Endpoint)

	tracerMu.Lock()
	defer tracerMu.Unlock()

	// Check if a tracer is already running for this endpoint
	if _, exists := activeTracers[key]; exists {
		return fmt.Errorf("tracer already exists for endpoint: %s", key)
	}

	// Initialize tracer provider with options
	opts, err := t.setTracerOptions()
	if err != nil {
		return err
	}

	t.tp = sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(t.tp)

	activeTracers[key] = struct{}{}

	return nil
}

// Stop stops the tracer.
func (t *tracer) Stop(ctx context.Context) error {
	key := fmt.Sprintf("%s:%s", t.config.Provider, t.config.Endpoint)

	tracerMu.Lock()
	defer tracerMu.Unlock()

	if t.tp != nil {
		err := t.tp.Shutdown(ctx)
		t.tp = nil
		// Cleanup tracer registration
		delete(activeTracers, key)
		return err
	}
	return nil
}

// GetTracerProvider returns the tracer provider for the tracer.
func (t *tracer) GetTracerProvider() *sdktrace.TracerProvider {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tp
}

// setTracerOptions sets the tracer options for the tracer provider.
func (t *tracer) setTracerOptions() ([]sdktrace.TracerProviderOption, error) {

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(t.config.SamplingRate))),
		sdktrace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(t.config.Name))),
	}

	exporter, err := t.createExporter()
	if err != nil {
		return nil, err
	}
	opts = append(opts, sdktrace.WithBatcher(exporter))

	return opts, nil
}

// createExporter creates a new opentelemetry span exporter.
func (t *tracer) createExporter() (sdktrace.SpanExporter, error) {
	if len(t.config.Endpoint) == 0 && t.config.Provider != ProviderStdout {
		return nil, errors.New("endpoint is required")
	}

	switch t.config.Provider {
	// DEPRECATED(hackerchai): Jeager is deprecated, please use otlp methods to connect to jaeger instead.
	case ProviderJaeger:
		u, err := url.Parse(t.config.Endpoint)
		if err == nil && u.Scheme == "udp" {
			return jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost(u.Hostname()),
				jaeger.WithAgentPort(u.Port())))
		}
		return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(t.config.Endpoint)))

	// Zipkin
	case ProviderZipkin:
		return zipkin.New(t.config.Endpoint)

	// Otlp is recommended for production environment.
	// Otlp Grpc
	case ProviderOtlpGrpc:
		return t.createOtlpGrpcExporter()

	// Otlp Http
	case ProviderOtlpHttp:
		return t.createOtlpHttpExporter()

	// File
	case ProviderFile:
		file, err := os.OpenFile(t.config.Endpoint, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		return stdouttrace.New(stdouttrace.WithWriter(file))

	// Stdout
	case ProviderStdout:
		return stdouttrace.New(stdouttrace.WithPrettyPrint())

	default:
		return nil, errors.New("unsupported provider: " + t.config.Provider)
	}
}

// createOtlpGrpcExporter creates a new otlp grpc exporter.
func (t *tracer) createOtlpGrpcExporter() (sdktrace.SpanExporter, error) {
	// Base options with endpoint
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(t.config.Endpoint),
	}

	// Map of boolean conditions to their corresponding options
	optionsMap := map[bool]otlptracegrpc.Option{
		!t.config.OtlpSecure:                   otlptracegrpc.WithInsecure(),
		t.config.OtlpRetryConfig.Enabled:       otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig(t.config.OtlpRetryConfig)),
		t.config.OtlpGrpcTLSCredentials != nil: otlptracegrpc.WithTLSCredentials(t.config.OtlpGrpcTLSCredentials),
		t.config.OtlpGrpcConn != nil:           otlptracegrpc.WithGRPCConn(t.config.OtlpGrpcConn),
	}

	// Add options based on conditions
	for condition, opt := range optionsMap {
		if condition {
			opts = append(opts, opt)
		}
	}

	// Handle non-zero duration configurations
	if t.config.OtlpTimeout > 0 {
		opts = append(opts, otlptracegrpc.WithTimeout(t.config.OtlpTimeout))
	}
	if t.config.OtlpGrpcReconnectionPeriod > 0 {
		opts = append(opts, otlptracegrpc.WithReconnectionPeriod(t.config.OtlpGrpcReconnectionPeriod))
	}

	// Handle non-empty slice/map configurations
	if len(t.config.OtlpHeaders) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(t.config.OtlpHeaders))
	}
	if len(t.config.OtlpGrpcServiceConfig) > 0 {
		opts = append(opts, otlptracegrpc.WithServiceConfig(t.config.OtlpGrpcServiceConfig))
	}

	// Handle compression settings
	var compressionOpt string
	switch t.config.OtlpCompression {
	case OtlpCompressionGzip:
		compressionOpt = "gzip"
	case OtlpCompressionNone:
		compressionOpt = ""
	default:
		compressionOpt = ""
	}
	opts = append(opts, otlptracegrpc.WithCompressor(compressionOpt))

	return otlptracegrpc.New(context.Background(), opts...)
}

// createOtlpHttpExporter creates a new otlp http exporter.
func (t *tracer) createOtlpHttpExporter() (sdktrace.SpanExporter, error) {
	// Base options with endpoint
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(t.config.Endpoint),
	}

	// Map of boolean conditions to their corresponding options
	optionsMap := map[bool]otlptracehttp.Option{
		!t.config.OtlpSecure:                    otlptracehttp.WithInsecure(),
		t.config.OtlpRetryConfig.Enabled:        otlptracehttp.WithRetry(otlptracehttp.RetryConfig(t.config.OtlpRetryConfig)),
		t.config.OtlpHttpTLSClientConfig != nil: otlptracehttp.WithTLSClientConfig(t.config.OtlpHttpTLSClientConfig),
	}

	// Add options based on conditions
	for condition, opt := range optionsMap {
		if condition {
			opts = append(opts, opt)
		}
	}

	// Handle timeout configuration
	if t.config.OtlpTimeout > 0 {
		opts = append(opts, otlptracehttp.WithTimeout(t.config.OtlpTimeout))
	}

	// Handle non-empty configurations
	if len(t.config.OtlpHeaders) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(t.config.OtlpHeaders))
	}
	if len(t.config.OtlpHttpPath) > 0 {
		opts = append(opts, otlptracehttp.WithURLPath(t.config.OtlpHttpPath))
	}

	// Handle compression settings
	switch t.config.OtlpCompression {
	case OtlpCompressionGzip:
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
	case OtlpCompressionNone:
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.NoCompression))
	default:
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.NoCompression))
	}

	return otlptracehttp.New(context.Background(), opts...)
}
