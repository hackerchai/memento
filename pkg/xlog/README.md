# xlog

A structured logging library built on top of [zerolog](https://github.com/rs/zerolog), providing enhanced tracing capabilities and context-aware logging. 

## Features

- Structured logging with JSON output
- Context-aware request ID tracking
- OpenTelemetry integration
- Configurable log levels and formats
- Log rotation support
- Thread-safe operations

## Installation

```bash
go get github.com/hackerchai/memento/pkg/xlog
```

## Quick Start

```go
logger := xlog.NewDefaultLogger()
logger.InfoMsg("Hello, World!")
```

## Basic Usage

### Creating a Logger

```go
// Default logger (outputs to stdout)
logger := xlog.NewDefaultLogger()

// Custom logger with specific configuration
logger := xlog.NewLogger(
    os.Stdout,
    xlog.Linfo,
    time.RFC3339,
    time.Local,
)
```

### Simple Logging

```go
// Log a simple trace message without structured logging
logger.TraceMsg("Service started")

// Log a fatal message with a formatted string
logger.FatalMsgf("This is a fatal message: %s", "create error")

// Log a simple message with structured logging
logger.DebugX().Str("key", "value").Msg("Hello, World!")
```

### Logging with Context

```go
ctx := context.Background()
logger := xlog.NewDefaultLogger()

// Add logger to context
ctx = logger.WithContext(ctx)

// Get logger from context
loggerFromCtx := xlog.Ctx(ctx)

// Log with context (includes trace information if tracing option is enabled)
loggerFromCtx.ErrorX(ctx).
    Str("key", "value").
    Msg("This is a contextual log message")
```

### Converting with existing zerolog

```go
// create a zerolog logger and convert it to xlog.Logger
zerologLogger := zerolog.New(os.Stdout).Logger()
logger := xlog.From(zerologLogger)

// get the underlying zerolog.Logger from xlog.Logger
zerologLogger, err := logger.GetLogger()
if err != nil {
    panic(err)
}

// if you ensure the logger is initialized, you can use Unwrap to get the underlying zerolog.Logger
zerologLogger := logger.Unwrap()
```

## Demo Examples

Check out the demo examples in the `_demo` directory:

### 1. Gin Integration (`_demo/gin/gin.go`)

Shows how to:

- Set up logging middleware in Gin
- Track request IDs across handlers using `context` to propagate request IDs
- Use `xlog.Ctx(ctx)` to get the logger from the context and pass it to services and repositories.

### 2. OpenTelemetry Integration (`_demo/trace/trace.go`)

Demonstrates:

- Trace context propagation using OpenTelemetry
- Span creation and management
- Low-invasive logging with trace information utilizing `context`

## Best Practices

1. **Context Usage**

   In a standard web service, we should pass and save `xlog.Logger` through the `Context` and pass the `Context` as a function parameter throughout the entire chain to the `Service` and `Repository`.

    ```go
    // Pass context through parameters
    func HandleRequest(ctx context.Context) {
        logger := xlog.Ctx(ctx)
        logger.InfoX(ctx).Msg("Processing request")
    }
    ```

2. **Middleware Integration**

    **Please do not use `xlog.New()`in middleware, it will affect performance. The same logging instance should be used to avoid memory allocation overhead.You can use `xlog.SetReqIDToContext()` to set the request ID to the context.**

    ```go
    func loggerMiddleware(logger *xlog.Logger) gin.HandlerFunc {
        return func(c *gin.Context) {
            reqID := xlog.GenReqId()
            ctx := logger.SetReqIDToContext(c.Request.Context(), reqID)
            c.Request = c.Request.WithContext(ctx)
            c.Next()
        }
    }
    ```

    If you want to add more fields like `path`, `method`, `status`, `latency`, etc., you can use `xlog.With()` to create a child logger. And pass the logger throughout the entire request chain.

    ```go
    func accesslogMiddleware(logger *xlog.Logger) gin.HandlerFunc { 
        return func(c *gin.Context) {
            childLogger := logger.With().Str("path", c.Request.URL.Path).Str("method", c.Request.Method).Logger()
            ctx := childLogger.WithContext(c.Request.Context())
            c.Request = c.Request.WithContext(ctx)
            c.Next()
        }
    }
    ```

3. **Avoid Using Global Logs**

   **Unless necessary, do not use global logs as they prevent the transmission of context information such as ReqID and Tracing, and they also affect the flexibility of log usage.**

   When needed, you can create a child logger using `xlog.With()` and attach the required fields for use. You can inject the logger into the required location via dependency injection.

   ```go
   type Service struct {
       logger *xlog.Logger
   }

   func NewService(logger *xlog.Logger) *Service {
       return &Service{
           logger: logger.With().Str("component", "Service").Logger(),
       }
   }

   func (s *Service) Process(ctx context.Context) {
       s.logger.InfoX(ctx).Msg("Processing started")
   }
   ```

4. **Log Rotation**

   xlog provides the `xlog.NewRotateWriter()` method to easily create a writer that supports log rotation, which can be directly passed into `xlog.NewLogger()` to create a logger that supports log rotation.

   ```go
   rotateWriter := xlog.NewRotateWriter(
       xlog.WithFilename("app.log"),
       xlog.WithMaxSize(100),
       xlog.WithMaxAge(7),
   )
   logger := xlog.NewLogger(rotateWriter, xlog.Linfo, "", nil)
   ```

## Configuration Methods

### Logger Settings

```go
logger := xlog.NewDefaultLogger()

// Set log level
logger.SetLevel(xlog.Ldebug)

// Set caller depth for stack traces
logger.SetCaller(3)

// Set time format and location
logger.SetTimeFormat(time.RFC3339)
logger.SetTimeLocation(time.UTC)

// Set field names
logger.SetTimeField("timestamp")
logger.SetLevelField("level")
logger.SetMessageField("message")
logger.SetCallerField("caller")
logger.SetErrorField("error")

// Set trace information
// Enable tracing, which will add traceID, spanID, reqID, etc. to the logs (if they exist)
// Enabled by default
logger.SetTrace(true)

// Set the key for ReqID in the Context
logger.SetReqIDKey("request_id")
```

### Custom Field Values

```go
// Customize level string representations
logger.SetLevelTraceValue("TRACE")
logger.SetLevelDebugValue("DEBUG")
logger.SetLevelInfoValue("INFO")
logger.SetLevelWarnValue("WARN")
logger.SetLevelErrorValue("ERROR")
logger.SetLevelFatalValue("FATAL")
logger.SetLevelPanicValue("PANIC")
```

### Error Handling

```go
import "github.com/rs/zerolog/pkgerrors"

// Set custom error stack field name
logger.SetErrorStackFieldName("stack_trace")

// Set the formatted error stack serialization method
logger.SetErrorStackMarshalFunc(pkgerrors.MarshalStack)

// Set custom error handler
logger.SetErrorHandler(func(err error) {
    fmt.Printf("Error occurred: %v\n", err)
})

// Set custom error marshaling
logger.SetErrorMarshalFunc(func(err error) interface{} {
    return err.Error()
})
```

### Beautify Terminal Output

xlog provides the `xlog.NewConsoleWriter()` method to conveniently create a writer that beautifies terminal output.

**It is not recommended to use a `ConsoleWriter` output in production environments, as `ConsoleWriter` can reduce log writing performance. Please use it in development environments.**

```go  
consoleWriter := xlog.NewConsoleWriter(os.Stdout, time.RFC3339, false)

logger := xlog.NewLogger(consoleWriter, xlog.Linfo, "", nil)
```

### Multiple Log Output

`xlog.NewMultiWriter()` may be used to send the log message to multiple outputs. We send the log message to both `os.Stdout` and the in-built ConsoleWriter for example.

```go
consoleWriter := xlog.NewConsoleWriter(os.Stdout, time.RFC3339, false)

multi := xlog.NewMultiWriter(consoleWriter, os.Stdout)

logger := xlog.NewLogger(multi, xlog.Linfo, "", nil)

logger.InfoMsg("Hello World!")

// Output (Line 1: Console; Line 2: Stdout)
// 12:36PM INF Hello World!
// {"level":"info","time":"2019-11-07T12:36:38+03:00","message":"Hello World!"}
```

## Sampling

xlog supports log sampling to reduce log volume in high-throughput scenarios:

### Basic Sampling

```go
// Sample every Nth message
sampler := xlog.NewBasicSampler(10) // Log every 10th message
logger.Sample(sampler)
```

### Level-Based Sampling

```go
// Will let 5 debug messages per period of 1 second.
// Over 5 debug message, 1 every 100 debug messages are logged.
// Other levels are not sampled.
sampled := log.Sample(xlog.LevelSampler{
    DebugSampler: &xlog.BurstSampler{
        Burst: 5,
        Period: 1*time.Second,
        NextSampler: &xlog.BasicSampler{N: 100},
    },
})
sampled.Debug().Msg("hello world")
```

## Hooks and Middleware

### Adding Context Information

```go
// Add custom fields to all log entries
// Note: This is not a thread-safe operation, please do not use it in concurrent operations.
logger.UpdateContext(func(c xlog.Context) xlog.Context {
    return c.Str("service", "api").
           Str("version", "1.0.0")
})
```

### Creating Child Loggers

```go
// Create a new logger with additional context
childLogger := logger.With().
    Str("component", "handler").
    Logger()

// Use in concurrent operations (thread-safe)
childLogger.InfoMsg("Handler started")

// Clone a logger
cloneLogger := logger.Clone()
```

### Custom Output with Hooks

```go
type SeverityHook struct{}

func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
    if level != zerolog.NoLevel {
        e.Str("severity", level.String())
    }
}

hooked := log.Hook(SeverityHook{})
hooked.Warn().Msg("")

// Output: {"level":"warn","severity":"warn"}
```

## Output Examples

### Standard Log Entry

```json
{
    "level": "info",
    "time": "2024-01-20T10:30:45Z",
    "caller": "app.go:42",
    "message": "Server started"
}
```

### Traced Log Entry

```json
{
    "level": "info",
    "time": "2024-01-20T10:30:45Z",
    "caller": "handler.go:24",
    "reqID": "7da6d442-1234-5678-90ab-cdef01234567",
    "traceID": "4bf92f3577b34da6a3ce929d0e0e4736",
    "spanID": "00f067aa0ba902b7",
    "message": "Request processed"
}
```

## Configuration Options

### Log Levels

- `Ltrace` Trace: The most detailed log level, typically used for debugging and development environments
- `Ldebug` Debug: Used for debugging and development environments
- `Linfo` Info: Used to record regular operations and important events
- `Lwarn` Warn: Used to record warning information
- `Lerror` Error: Used to record errors and exceptions
- `Lfatal` Fatal: Used to record critical errors and program termination
- `Lpanic` Panic: Used to record program crashes

### Time Formats

- `TimeFormatUnix` Timestamp
- `TimeFormatUnixMs` Timestamp (milliseconds)
- `TimeFormatUnixMicro` Timestamp (microseconds)
- `TimeFormatUnixNano` Timestamp (nanoseconds)

## Thread Safety

The logger is safe for concurrent use. However, be careful with `UpdateContext` - use `With()` to create child loggers for concurrent operations
