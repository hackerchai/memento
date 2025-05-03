# xlog

一个基于 [zerolog](https://github.com/rs/zerolog) 构建的结构化日志库，提供增强的跟踪功能和上下文感知的日志记录

## 功能特色

- 结构化日志输出为JSON格式
- 上下文感知的请求ID跟踪
- OpenTelemetry集成
- 可配置的日志级别和格式
- 日志轮转支持
- 线程安全的操作

## 安装

```bash
go get github.com/hackerchai/memento/pkg/xlog
```

## 快速开始

```go
logger := xlog.NewDefaultLogger()
logger.InfoMsg("Hello, World!")
```

## 基本用法

### 创建一个Logger

```go
// 默认 logger（输出到 stdout）
logger := xlog.NewDefaultLogger()

// 具有特定配置的自定义 logger
logger := xlog.NewLogger(
    os.Stdout,
    xlog.Linfo,
    time.RFC3339,
    time.Local,
)
```

### 简单日志记录

```go
// 记录一个简单的跟踪消息，不带结构化日志
logger.TraceMsg("Service started")

// 使用格式化字符串记录一个致命错误消息
logger.FatalMsgf("This is a fatal message: %s", "create error")

// 记录一个带结构化日志的简单消息
logger.DebugX().Str("key", "value").Msg("Hello, World!")
```

### 使用上下文记录日志

```go
ctx := context.Background()
logger := xlog.NewDefaultLogger()

// 将 logger 添加到上下文
ctx = logger.WithContext(ctx)

// 从上下文中获取 logger
loggerFromCtx := xlog.Ctx(ctx)

// 使用上下文记录日志（如果跟踪选项启用，包含跟踪信息）
loggerFromCtx.ErrorX(ctx).
    Str("key", "value").
    Msg("This is a contextual log message")

```

### 与现有的 zerolog 进行转换

```go
// 创建一个 zerolog logger 并将其转换为xlog.Logger
zerologLogger := zerolog.New(os.Stdout).Logger()
logger := xlog.From(zerologLogger)

// 从 xlog.Logger 中获取底层的 zerolog.Logger
zerologLogger, err := logger.GetLogger()
if err != nil {
    panic(err)
}

// 如果你确保logger已初始化，可以使用 Unwrap 方法获取底层的 zerolog.Logger
zerologLogger := logger.Unwrap()
```

## 演示示例

查看`_demo`目录中的演示示例：

### 1. Gin 集成（`_demo/gin/gin.go`）

演示如何：

- 在 Gin 中设置日志中间件
- 使用 `context` 在处理程序之间传递 ReqID
- 使用 `xlog.Ctx(ctx)` 从上下文中获取 logger 并将其传递给服务和存储库。

### 2. OpenTelemetry 集成（`_demo/trace/trace.go`）

展示：

- 使用 OpenTelemetry 进行跟踪上下文传播
- Span 的创建和管理
- 利用 `Context` 进行低侵入的日志记录和跟踪信息

## 最佳实践

1. **上下文使用**

   在标准的 Web 服务中，我们应当通过 `Context` 上下文来传递和保存 xlog.Logger，并且将 Context 使用函数参数的方式在整个链路中传递给 `Service` 和 `Repository`。

    ```go
    // 通过 Context 上下文来传递 xlog.Logger
    func HandleRequest(ctx context.Context) {
        logger := xlog.Ctx(ctx)
        logger.InfoX(ctx).Msg("Processing request")
    }
    ```

2. **中间件集成**

    **请勿在中间件中频繁使用`xlog.New()`，这会影响性能。应当使用同一个日志实例以避免内存分配损耗。你可以使用 `xlog.SetReqIDToContext()` 将 ReqID 设置到上下文中。**

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

    如果你想添加如 `path`、`method`、`status`、`latency` 等字段，可以使用 `xlog.With()` 创建一个子 logger，并将 logger 传递给整个请求链路。

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

3. **避免使用全局日志**

    **如非必要请勿使用全局日志，这会导致日志信息无法传递 ReqID 和 Tracing 等上下文信息，也会影响日志使用的灵活性**

    在需要时，可以通过 `xlog.With()` 创建一个子 logger 并且附加所需要的字段以供使用，可以通过依赖注入的方式将 logger 注入到所需要的位置

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

4. **日志轮转**

    xlog 提供了 `xlog.NewRotateWriter()` 方法来方便的创建一个支持日志轮转的 writer，可以直接传入 `xlog.NewLogger()` 来创建一个支持日志轮转的 logger

    ```go
    rotateWriter := xlog.NewRotateWriter(
        xlog.WithFilename("app.log"),
        xlog.WithMaxSize(100),
        xlog.WithMaxAge(7),
    )
    logger := xlog.NewLogger(rotateWriter, xlog.Linfo, "", nil)
    ```

## 配置方法

### Logger设置

```go
logger := xlog.NewDefaultLogger()

// 设置日志级别
logger.SetLevel(xlog.Ldebug)

// 设置堆栈跟踪的调用者深度
logger.SetCaller(3)

// 设置时间格式和时区
logger.SetTimeFormat(time.RFC3339)
logger.SetTimeLocation(time.UTC)

// 设置字段名称
logger.SetTimeField("timestamp")
logger.SetLevelField("level")
logger.SetMessageField("message")
logger.SetCallerField("caller")
logger.SetErrorField("error")

// 设置跟踪信息
// 开启跟踪，开启后会在日志中添加 traceID、spanID、reqID 等信息(若存在)
// 默认开启
logger.SetTrace(true)

// 设置 ReqID 在 Context 中的键值
logger.SetReqIDKey("request_id")
```

### 自定义字段值

```go
// 自定义日志级别字符串表示
logger.SetLevelTraceValue("TRACE")
logger.SetLevelDebugValue("DEBUG")
logger.SetLevelInfoValue("INFO")
logger.SetLevelWarnValue("WARN")
logger.SetLevelErrorValue("ERROR")
logger.SetLevelFatalValue("FATAL")
logger.SetLevelPanicValue("PANIC")
```

### 错误处理

```go
import "github.com/rs/zerolog/pkgerrors"

// 设置自定义错误堆栈字段名称
logger.SetErrorStackFieldName("stack_trace")

// 设置格式化的错误堆栈序列化方法
logger.SetErrorStackMarshalFunc(pkgerrors.MarshalStack)

// 设置自定义错误处理器
logger.SetErrorHandler(func(err error) {
    fmt.Printf("Error occurred: %v\n", err)
})

// 设置自定义错误序列化
logger.SetErrorMarshalFunc(func(err error) interface{} {
    return err.Error()
})
```

### 美化终端输出

xlog 提供了 `xlog.NewConsoleWriter()` 方法来方便的创建一个美化终端输出的 writer

**不推荐在生产环境中使用 `ConsoleWriter` ，因为 `ConsoleWriter` 会降低日志写入性能，请在开发环境中使用**

```go  
consoleWriter := xlog.NewConsoleWriter(os.Stdout, time.RFC3339, false)

logger := xlog.NewLogger(consoleWriter, xlog.Linfo, "", nil)
```

### 多输出日志

`xlog.NewMultiWriter()` 可以用于将日志消息发送到多个输出中。这里我们通过一个示例将日志消息发送到 `os.Stdout` 和内置的 ConsoleWriter。

```go
consoleWriter := xlog.NewConsoleWriter(os.Stdout, time.RFC3339, false)

multi := xlog.NewMultiWriter(consoleWriter, os.Stdout)

logger := xlog.NewLogger(multi, xlog.Linfo, "", nil)

logger.InfoMsg("Hello World!")

// 输出 (第1行: 控制台; 第2行: 标准输出)
// 12:36PM INF Hello World!
// {"level":"info","time":"2019-11-07T12:36:38+03:00","message":"Hello World!"}
```

## 采样

xlog 支持日志采样，以减少高吞吐量场景下的日志量：

### 基本采样

```go
// 每 Nth 条消息采样
sampler := xlog.NewBasicSampler(10) // 每 10 条消息记录一条日志
logger.Sample(sampler)
```

### 基于级别的采样

```go
// 每秒记录 5 条 debug 消息。超过 5 条后，每 100 条消息记录 1 条
// 其他级别不采样
sampled := log.Sample(xlog.LevelSampler{
    DebugSampler: &xlog.BurstSampler{
        Burst: 5,
        Period: 1*time.Second,
        NextSampler: &xlog.BasicSampler{N: 100},
    },
})
sampled.Debug().Msg("hello world")
```

## 钩子和中间件

### 添加上下文信息

```go
// 为所有日志条目添加自定义字段
logger.UpdateContext(func(c xlog.Context) xlog.Context {
    return c.Str("service", "api").
           Str("version", "1.0.0")
})
```

### 创建子Loggers

```go
// 创建一个具有附加上下文的新 logger
childLogger := logger.With().
    Str("component", "handler").
    Logger()

// 在并发操作中使用（线程安全）
childLogger.InfoMsg("Handler started")

// 克隆一个 logger
cloneLogger := logger.Clone()
```

### 使用钩子自定义输出

```go
type SeverityHook struct{}

func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
    if level != zerolog.NoLevel {
        e.Str("severity", level.String())
    }
}

hooked := log.Hook(SeverityHook{})
hooked.Warn().Msg("")

// 输出: {"level":"warn","severity":"warn"}
```

## 输出示例

### 标准日志条目

```json
{
    "level": "info",
    "time": "2024-01-20T10:30:45Z",
    "caller": "app.go:42",
    "message": "Server started"
}
```

### 带跟踪的日志条目

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

## 配置选项

### 日志级别

- `Ltrace` 跟踪: 最详细的日志级别，通常用于调试和开发环境
- `Ldebug` 调试: 用于调试和开发环境
- `Linfo` 信息: 用于记录常规操作和重要事件
- `Lwarn` 警告: 用于记录警告信息
- `Lerror` 错误: 用于记录错误和异常
- `Lfatal` 严重错误: 用于记录严重错误和程序终止
- `Lpanic` 崩溃: 用于记录程序崩溃

### 时间格式

- `TimeFormatUnix` 时间戳
- `TimeFormatUnixMs` 时间戳（毫秒）
- `TimeFormatUnixMicro` 时间戳（微秒）
- `TimeFormatUnixNano` 时间戳（纳秒）

## 线程安全

Logger 可以安全地用于并发使用。然而，在使用 `UpdateContext` 时需小心 - 使用 `With()` 为并发操作创建子 logger
