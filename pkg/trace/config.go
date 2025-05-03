package trace

import (
	"crypto/tls"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// OtlpCompression defines the compression type for OTLP
type OtlpCompression string

const (
	OtlpCompressionNone OtlpCompression = "none"
	OtlpCompressionGzip OtlpCompression = "gzip"
)

type RetryConfig otlptracegrpc.RetryConfig

type Config struct {
	// Name is the name of the tracer.
	Name string `json:"name" yaml:"name"`
	// Endpoint is the endpoint for the tracer.
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	// SamplingRate is the sampling rate for the tracer.
	// Default is 1.0, which means 100% of the traces are sampled.
	SamplingRate float64 `json:"sampling_rate" yaml:"sampling_rate"`
	// Provider is the provider for the tracer.
	// Accepted values: ProviderOtlpGrpc, ProviderOtlpHttp, ProviderFile, ProviderStdout, ProviderJaeger, ProviderZipkin
	Provider string `json:"provider" yaml:"provider"`
	// OtlpHeaders is the headers for the otlp exporter.
	OtlpHeaders map[string]string `json:"otlp_headers" yaml:"otlp_headers"`
	// OtlpSecure is the secure flag for the otlp exporter to choose between http and https.
	OtlpSecure bool `json:"otlp_secure" yaml:"otlp_secure"`
	// OtlpCompression is the compression for the otlp exporter.
	// Accepted values: OtlpCompressionGzip, OtlpCompressionNone
	OtlpCompression OtlpCompression `json:"otlp_compression" yaml:"otlp_compression"`
	// OtlpTimeout is the timeout for the otlp exporter.
	OtlpTimeout time.Duration `json:"otlp_timeout" yaml:"otlp_timeout"`
	// OtlpRetryConfig is the retry config for the otlp exporter.
	OtlpRetryConfig RetryConfig `json:"otlp_retry_config" yaml:"otlp_retry_config"`
	// OtlpGrpcServiceConfig is the service config for the otlp grpc exporter.
	OtlpGrpcServiceConfig string `json:"otlp_grpc_service_config" yaml:"otlp_grpc_service_config"`
	// OtlpGrpcDialOptions is the dial options for the otlp grpc exporter.
	OtlpGrpcDialOptions []grpc.DialOption `json:"otlp_grpc_dial_options" yaml:"otlp_grpc_dial_options"`
	// OtlpGrpcTLSCredentials is the tls credentials for the otlp grpc exporter.
	OtlpGrpcTLSCredentials credentials.TransportCredentials `json:"otlp_grpc_tls_credentials" yaml:"otlp_grpc_tls_credentials"`
	// OtlpGrpcReconnectionPeriod is the reconnection period for the otlp grpc exporter.
	OtlpGrpcReconnectionPeriod time.Duration `json:"otlp_grpc_reconnection_period" yaml:"otlp_grpc_reconnection_period"`
	// OtlpGrpcConn is the grpc client connection for the otlp grpc exporter.
	OtlpGrpcConn *grpc.ClientConn `json:"otlp_grpc_conn" yaml:"otlp_grpc_conn"`
	// OtlpHttpPath is the path for the otlp http exporter.
	OtlpHttpPath string `json:"otlp_http_path" yaml:"otlp_http_path"`
	// OtlpHttpTLSClientConfig is the tls client config for the otlp http exporter.
	OtlpHttpTLSClientConfig *tls.Config `json:"otlp_http_tls_client_config" yaml:"otlp_http_tls_client_config"`
}

// NewTracerConfig creates a new Config with default values
func NewTracerConfig() *Config {
	return &Config{
		Name:         "default-tracer",
		SamplingRate: 1.0,
		Provider:     ProviderOtlpGrpc,
	}
}

// WithName sets the tracer name
func (c *Config) WithName(name string) *Config {
	c.Name = name
	return c
}

// WithEndpoint sets the endpoint for the tracer
func (c *Config) WithEndpoint(endpoint string) *Config {
	c.Endpoint = endpoint
	return c
}

// WithSamplingRate sets the sampling rate for the tracer
func (c *Config) WithSamplingRate(rate float64) *Config {
	c.SamplingRate = rate
	return c
}

// WithProvider sets the provider for the tracer
func (c *Config) WithProvider(provider string) *Config {
	c.Provider = provider
	return c
}

// WithOtlpHeaders sets the headers for the OTLP exporter
func (c *Config) WithOtlpHeaders(headers map[string]string) *Config {
	c.OtlpHeaders = headers
	return c
}

// WithOtlpSecure sets the secure flag for the OTLP exporter
func (c *Config) WithOtlpSecure(secure bool) *Config {
	c.OtlpSecure = secure
	return c
}

// WithOtlpCompression sets the compression for the OTLP exporter
func (c *Config) WithOtlpCompression(compression OtlpCompression) *Config {
	c.OtlpCompression = compression
	return c
}

// WithOtlpTimeout sets the timeout for the OTLP exporter
func (c *Config) WithOtlpTimeout(timeout time.Duration) *Config {
	c.OtlpTimeout = timeout
	return c
}

// WithOtlpRetryConfig sets the retry config for the OTLP exporter
func (c *Config) WithOtlpRetryConfig(retryConfig RetryConfig) *Config {
	c.OtlpRetryConfig = retryConfig
	return c
}

// WithOtlpGrpcServiceConfig sets the service config for the OTLP gRPC exporter
func (c *Config) WithOtlpGrpcServiceConfig(serviceConfig string) *Config {
	c.OtlpGrpcServiceConfig = serviceConfig
	return c
}

// WithOtlpGrpcDialOptions sets the dial options for the OTLP gRPC exporter
func (c *Config) WithOtlpGrpcDialOptions(dialOptions []grpc.DialOption) *Config {
	c.OtlpGrpcDialOptions = dialOptions
	return c
}

// WithOtlpGrpcTLSCredentials sets the TLS credentials for the OTLP gRPC exporter
func (c *Config) WithOtlpGrpcTLSCredentials(creds credentials.TransportCredentials) *Config {
	c.OtlpGrpcTLSCredentials = creds
	return c
}

// WithOtlpGrpcReconnectionPeriod sets the reconnection period for the OTLP gRPC exporter
func (c *Config) WithOtlpGrpcReconnectionPeriod(period time.Duration) *Config {
	c.OtlpGrpcReconnectionPeriod = period
	return c
}

// WithOtlpGrpcConn sets the gRPC client connection for the OTLP gRPC exporter
func (c *Config) WithOtlpGrpcConn(conn *grpc.ClientConn) *Config {
	c.OtlpGrpcConn = conn
	return c
}

// WithOtlpHttpPath sets the path for the OTLP HTTP exporter
func (c *Config) WithOtlpHttpPath(path string) *Config {
	c.OtlpHttpPath = path
	return c
}

// WithOtlpHttpTLSClientConfig sets the TLS client config for the OTLP HTTP exporter
func (c *Config) WithOtlpHttpTLSClientConfig(tlsConfig *tls.Config) *Config {
	c.OtlpHttpTLSClientConfig = tlsConfig
	return c
}
