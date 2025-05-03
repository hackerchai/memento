# Trace Link Tracking Component

This is a link tracking component based on OpenTelemetry, supporting multiple exporters and providing integration support with the Gin framework.

## Features

- Supports multiple tracing data export methods:
  - OTLP (gRPC/HTTP)
  - Jaeger (deprecated, OTLP is recommended)
  - Zipkin
  - File output
  - Standard output
- Provides seamless integration with the Gin framework
- Supports flexible sampling rate configuration
- Supports TLS secure transmission
- Supports request header filtering
- Supports path filtering

## Installation

```bash
go get github.com/hackerchai/memento/pkg/trace
```

## Basic Usage

### 1. Create a Tracer

```go
// Create basic configuration
config := trace.NewTracerConfig().
    WithName("my-service").
    WithProvider(trace.ProviderOtlpGrpc).
    WithEndpoint("localhost:4317").
    WithSamplingRate(1.0)

// OTLP GRPC specific configuration
config.
    WithOtlpHeaders(map[string]string{
        "authorization": "bearer token",
    }).
    WithOtlpSecure(true).
    WithOtlpCompression(trace.OtlpCompressionGzip).
    WithOtlpTimeout(time.Second * 5).
    WithOtlpRetryConfig(trace.RetryConfig{
        Enabled: true,
        // ... other retry parameters
    }).
    WithOtlpGrpcServiceConfig("...").
    WithOtlpGrpcDialOptions([]grpc.DialOption{}).
    WithOtlpGrpcTLSCredentials(credentials.TransportCredentials).
    WithOtlpGrpcReconnectionPeriod(time.Second * 30).
    WithOtlpGrpcConn(grpcConn)

// OTLP HTTP specific configuration
config.
    WithOtlpHttpPath("/v1/traces").
    WithOtlpHttpTLSClientConfig(&tls.Config{})

// Initialize the tracer
tracer := trace.NewTracer(config)
```

### 2. Use the Tracer

```go
// Start the tracer
if err := tracer.Start(); err != nil {
    log.Fatal(err)
}
// Stop the tracer
defer tracer.Stop(context.Background())

// Get TracerProvider
tp := tracer.GetTracerProvider()
```

### 3. Integration with the Gin Framework

Here we use the otelgin library as a middleware for the Gin framework. For more configuration, please refer to [otelgin](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/github.com/gin-gonic/gin/otelgin)

```go
// Create basic configuration and initialize tracer
config := trace.NewTracerConfig().
    WithName("my-service").
    WithProvider(trace.ProviderOtlpGrpc).
    WithEndpoint("localhost:4317")

tracer := trace.NewTracer(config)

// Get TracerProvider
tp := tracer.GetTracerProvider()

// Create Gin instance
r := gin.New()

// Configure Gin middleware options
opts := []otelgin.Option{
    // Use our TracerProvider
    otelgin.WithTracerProvider(tp),
    
    // Set request filter
    otelgin.WithFilter(func(r *http.Request) bool {
        return !strings.Contains(r.URL.Path, "/health")
    }),
    
    // Set filter based on Gin Context
    otelgin.WithGinFilter(func(c *gin.Context) bool {
        return !strings.HasPrefix(c.FullPath(), "/internal")
    }),
    
    // Custom Span name formatting
    otelgin.WithSpanNameFormatter(func(r *http.Request) string {
        return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
    }),
}

// Register middleware, set service name
r.Use(otelgin.Middleware("my-service", opts...))

// Register routes
r.GET("/api/v1/users", func(c *gin.Context) {
    // Business logic...
    c.JSON(200, gin.H{"message": "success"})
})
```

## Best Practices

1. **Set a reasonable sampling rate**

   ```go
   config.WithSamplingRate(0.1) // Sample 10% of requests
   ```

2. **Filter sensitive information**

   ```go
   ginConfig.Filters = append(ginConfig.Filters,
       func(key string, value interface{}) bool {
           sensitiveHeaders := []string{
               "authorization",
               "cookie",
               "x-api-key",
           }
           for _, h := range sensitiveHeaders {
               if key == "http.request.header."+h {
                   return false
               }
           }
           return true
       },
   )
   ```

3. **Skip health checks and other paths**

   ```go
   ginConfig.SkipPaths = []string{
       "/health",
       "/metrics",
       "/favicon.ico",
   }
   ```

4. **Use OpenTelemetry Collector to export data**

    Use the [opentelemetry-collector](https://github.com/open-telemetry/opentelemetry-collector) to export data as an intermediary layer for exporters. It allows more flexible Sampling strategy configuration and more efficient data processing, eventually exporting to Jaeger or other backends.

    For configuration, refer to the official documentation [opentelemetry-collector-configuration](https://opentelemetry.io/docs/collector/configuration/)

## Notes

1. The Jaeger exporter has been marked as deprecated; it is recommended to use the OTLP exporter instead.
2. In production environments, it is recommended to use the OTLP protocol to connect to observability backends.
3. Please configure the sampling rate reasonably to avoid generating excessive data.
4. Ensure proper handling of the start and stop lifecycle of the tracer.

## Example Code

Complete example code can be found in `_demo/gin.go`
