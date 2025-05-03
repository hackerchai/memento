package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"

	"github.com/hackerchai/memento/pkg/trace"
)

// GinTracerConfig holds configuration for the Gin tracer middleware
type GinTracerConfig struct {
	// ServiceName is used to identify the service in traces
	ServiceName string
	// SkipPaths are the paths that won't be traced
	SkipPaths []string
	// Tracer is the tracer to use
	Tracer trace.Tracer
	// Filters for modifying span attributes
	Filters []otelgin.Filter
}

// NewDefaultGinConfig creates a new GinTracerConfig with default values
func NewDefaultGinConfig(serviceName string, tracerProvider trace.Tracer) *GinTracerConfig {
	return &GinTracerConfig{
		ServiceName: serviceName,
		Tracer:      tracerProvider,
		SkipPaths:   []string{"/health", "/metrics"},
		Filters:     []otelgin.Filter{},
	}
}

func StartGinServer(config *GinTracerConfig) error {
	if err := config.Tracer.Start(); err != nil {
		return err
	}
	defer config.Tracer.Stop(context.Background())

	// Create Gin engine
	r := gin.New()

	// Configure otelgin middleware with the provided config
	middlewareOpts := []otelgin.Option{
		otelgin.WithTracerProvider(config.Tracer.GetTracerProvider()),
		otelgin.WithFilter(config.Filters...),
	}

	// Skip paths if configured
	if len(config.SkipPaths) > 0 {
		middlewareOpts = append(middlewareOpts, otelgin.WithSkipPaths(config.SkipPaths...))
	}

	// Add middleware
	r.Use(otelgin.Middleware(config.ServiceName, middlewareOpts...))

	// Add example routes
	r.GET("/hello", func(c *gin.Context) {
		ctx := c.Request.Context()
		span := otel.Tracer(config.ServiceName).Start(ctx, "hello-operation")
		defer span.End()

		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, OpenTelemetry!",
		})
	})

	log.Println("Starting server on :8080")
	return r.Run(":8080")
}

func main() {
	// Create tracer
	tracerConfig := trace.NewTracerConfig().
		WithName("gin-demo").
		WithProvider(trace.ProviderStdout)

	tracer := trace.NewTracer(tracerConfig)

	// Create gin config
	ginConfig := NewDefaultGinConfig("gin-server", tracer)

	// Add custom filters if needed
	ginConfig.Filters = append(ginConfig.Filters,
		func(key string, value interface{}) bool {
			// Filter out sensitive information
			return key != "http.request.header.authorization"
		},
	)

	// Add more paths to skip if needed
	ginConfig.SkipPaths = append(ginConfig.SkipPaths, "/favicon.ico")

	// Start server
	if err := StartGinServer(ginConfig); err != nil {
		log.Fatal(err)
	}
}
