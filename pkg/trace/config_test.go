package trace

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewConfig(t *testing.T) {
	cfg := NewTracerConfig()
	assert.Equal(t, "default-tracer", cfg.Name)
	assert.Equal(t, 1.0, cfg.SamplingRate)
	assert.Equal(t, ProviderOtlpGrpc, cfg.Provider)
}

func TestConfig_WithMethods(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Config) *Config
		validate func(*testing.T, *Config)
	}{
		{
			name: "WithName",
			setup: func(c *Config) *Config {
				return c.WithName("test-tracer")
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, "test-tracer", c.Name)
			},
		},
		{
			name: "WithEndpoint",
			setup: func(c *Config) *Config {
				return c.WithEndpoint("localhost:4317")
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, "localhost:4317", c.Endpoint)
			},
		},
		{
			name: "WithSamplingRate",
			setup: func(c *Config) *Config {
				return c.WithSamplingRate(0.5)
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, 0.5, c.SamplingRate)
			},
		},
		{
			name: "WithProvider",
			setup: func(c *Config) *Config {
				return c.WithProvider(ProviderOtlpHttp)
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, ProviderOtlpHttp, c.Provider)
			},
		},
		{
			name: "WithOtlpHeaders",
			setup: func(c *Config) *Config {
				headers := map[string]string{"key": "value"}
				return c.WithOtlpHeaders(headers)
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, map[string]string{"key": "value"}, c.OtlpHeaders)
			},
		},
		{
			name: "WithOtlpSecure",
			setup: func(c *Config) *Config {
				return c.WithOtlpSecure(true)
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.True(t, c.OtlpSecure)
			},
		},
		{
			name: "WithOtlpTimeout",
			setup: func(c *Config) *Config {
				return c.WithOtlpTimeout(5 * time.Second)
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, 5*time.Second, c.OtlpTimeout)
			},
		},
		{
			name: "WithOtlpHttpPath",
			setup: func(c *Config) *Config {
				return c.WithOtlpHttpPath("/v1/traces")
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, "/v1/traces", c.OtlpHttpPath)
			},
		},
		{
			name: "WithOtlpHttpTLSClientConfig",
			setup: func(c *Config) *Config {
				tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
				return c.WithOtlpHttpTLSClientConfig(tlsConfig)
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.NotNil(t, c.OtlpHttpTLSClientConfig)
				assert.Equal(t, uint16(tls.VersionTLS12), c.OtlpHttpTLSClientConfig.MinVersion)
			},
		},
		{
			name: "WithOtlpGrpcServiceConfig",
			setup: func(c *Config) *Config {
				return c.WithOtlpGrpcServiceConfig("test-service")
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, "test-service", c.OtlpGrpcServiceConfig)
			},
		},
		{
			name: "WithOtlpGrpcDialOptions",
			setup: func(c *Config) *Config {
				return c.WithOtlpGrpcDialOptions([]grpc.DialOption{grpc.WithBlock()})
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Len(t, c.OtlpGrpcDialOptions, 1)
			},
		},
		{
			name: "WithOtlpGrpcTLSCredentials",
			setup: func(c *Config) *Config {
				return c.WithOtlpGrpcTLSCredentials(credentials.NewTLS(nil))
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.NotNil(t, c.OtlpGrpcTLSCredentials)
			},
		},
		{
			name: "WithOtlpGrpcReconnectionPeriod",
			setup: func(c *Config) *Config {
				return c.WithOtlpGrpcReconnectionPeriod(5 * time.Second)
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, 5*time.Second, c.OtlpGrpcReconnectionPeriod)
			},
		},
		{
			name: "WithOtlpCompression",
			setup: func(c *Config) *Config {
				return c.WithOtlpCompression(OtlpCompressionGzip)
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, OtlpCompressionGzip, c.OtlpCompression)
			},
		},
		{
			name: "WithOtlpRetry",
			setup: func(c *Config) *Config {
				return c.WithOtlpRetryConfig(RetryConfig{
					Enabled:         true,
					InitialInterval: 1 * time.Second,
					MaxInterval:     30 * time.Second,
					MaxElapsedTime:  time.Minute,
				})
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, RetryConfig{
					Enabled:         true,
					InitialInterval: 1 * time.Second,
					MaxInterval:     30 * time.Second,
					MaxElapsedTime:  time.Minute,
				}, c.OtlpRetryConfig)
			},
		},
		{
			name: "WithGrpcConn",
			setup: func(c *Config) *Config {
				conn, err := grpc.Dial("localhost:4317", grpc.WithTransportCredentials(insecure.NewCredentials()))
				assert.NoError(t, err)
				return c.WithOtlpGrpcConn(conn)
			},
			validate: func(t *testing.T, c *Config) {
				t.Helper()
				assert.NotNil(t, c.OtlpGrpcConn)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTracerConfig()
			cfg = tt.setup(cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestConfig_ChainedWithMethods(t *testing.T) {
	cfg := NewTracerConfig().
		WithName("chain-test").
		WithEndpoint("localhost:4317").
		WithSamplingRate(0.5).
		WithProvider(ProviderOtlpHttp).
		WithOtlpSecure(true)

	assert.Equal(t, "chain-test", cfg.Name)
	assert.Equal(t, "localhost:4317", cfg.Endpoint)
	assert.Equal(t, 0.5, cfg.SamplingRate)
	assert.Equal(t, ProviderOtlpHttp, cfg.Provider)
	assert.True(t, cfg.OtlpSecure)
}
