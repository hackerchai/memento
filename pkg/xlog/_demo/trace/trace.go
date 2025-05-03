package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/hackerchai/memento/pkg/xlog"
)

const serviceName = "demo-service"

// initTracer initializes the OpenTelemetry tracer
func initTracer() (*sdktrace.TracerProvider, error) {
	// Create a stdout exporter for demonstration purposes
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	// Create a new trace provider with the exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

// traceMiddleware is a Gin middleware that creates trace spans and injects request IDs
func traceMiddleware(logger *xlog.Logger) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)

	return func(c *gin.Context) {
		// Generate request ID using xlog's built-in method
		reqID := xlog.GenReqId()
		ctx := logger.SetReqIDToContext(c.Request.Context(), reqID)

		// Create a new span for this request
		spanCtx, span := tracer.Start(ctx, c.Request.URL.Path)
		defer span.End()

		// Store the span context in gin context
		c.Request = c.Request.WithContext(spanCtx)

		c.Next()
	}
}

func main() {
	// Initialize the tracer
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Initialize the logger
	logger := xlog.NewDefaultLogger()

	// Create Gin engine with the trace middleware
	r := gin.New()
	r.Use(traceMiddleware(logger))

	// Example route handler
	r.GET("/hello", func(c *gin.Context) {
		ctx := c.Request.Context()

		// Create a child span for this operation
		_, span := otel.Tracer(serviceName).Start(ctx, "hello-operation")
		defer span.End()

		// Log request information with trace context
		logger.InfoX(ctx).
			Str("path", c.Request.URL.Path).
			Msg("Handling hello request")
		// Example log output:
		// {
		//   "level": "info",
		//   "time": "2024-01-20T10:30:45Z",
		//   "caller": "trace.go:89",
		//   "reqID": "7da6d442-1234-5678-90ab-cdef01234567",
		//   "traceID": "4bf92f3577b34da6a3ce929d0e0e4736",
		//   "spanID": "00f067aa0ba902b7",
		//   "sampled": true,
		//   "path": "/hello",
		//   "message": "Handling hello request"
		// }

		// Simulate some business logic
		handleBusinessLogic(ctx, logger)

		c.JSON(200, gin.H{"message": "Hello, World!"})
	})

	r.Run(":8080")
}

// handleBusinessLogic simulates processing business logic with tracing
func handleBusinessLogic(ctx context.Context, logger *xlog.Logger) {
	// Create a child span for business logic
	_, span := otel.Tracer(serviceName).Start(ctx, "business-logic")
	defer span.End()

	// Log business operation
	logger.InfoX(ctx).
		Str("operation", "business-logic").
		Msg("Processing business logic")
	// Example log output:
	// {
	//   "level": "info",
	//   "time": "2024-01-20T10:30:45Z",
	//   "caller": "trace.go:120",
	//   "reqID": "7da6d442-1234-5678-90ab-cdef01234567",
	//   "traceID": "4bf92f3577b34da6a3ce929d0e0e4736",
	//   "spanID": "abc067aa0ba902de",
	//   "sampled": true,
	//   "operation": "business-logic",
	//   "message": "Processing business logic"
	// }

	// Simulate logging an error
	logger.ErrorX(ctx).
		Str("error_type", "business_validation").
		Msg("Something went wrong in business logic")
	// Example log output:
	// {
	//   "level": "error",
	//   "time": "2024-01-20T10:30:45Z",
	//   "caller": "trace.go:137",
	//   "reqID": "7da6d442-1234-5678-90ab-cdef01234567",
	//   "traceID": "4bf92f3577b34da6a3ce929d0e0e4736",
	//   "spanID": "abc067aa0ba902de",
	//   "sampled": true,
	//   "error_type": "business_validation",
	//   "message": "Something went wrong in business logic"
	// }
}
