# Trace 链路追踪组件

这是一个基于 OpenTelemetry 的链路追踪组件，支持多种导出器，并提供了与 Gin 框架的集成支持。

## 功能特性

- 支持多种追踪数据导出方式:
  - OTLP (gRPC/HTTP)
  - Jaeger (已弃用，建议使用 OTLP)
  - Zipkin
  - 文件输出
  - 标准输出
- 提供与 Gin 框架的无缝集成
- 支持灵活的采样率配置
- 支持 TLS 安全传输
- 支持请求头过滤
- 支持路径过滤

## 安装

```bash
go get github.com/hackerchai/memento/pkg/trace
```

## 基础使用

### 1. 创建追踪器

```go
// 创建基础配置
config := trace.NewTracerConfig().
    WithName("my-service").
    WithProvider(trace.ProviderOtlpGrpc).
    WithEndpoint("localhost:4317").
    WithSamplingRate(1.0)

// OTLP GRPC 专用配置
config.
    WithOtlpHeaders(map[string]string{
        "authorization": "bearer token",
    }).
    WithOtlpSecure(true).
    WithOtlpCompression(trace.OtlpCompressionGzip).
    WithOtlpTimeout(time.Second * 5).
    WithOtlpRetryConfig(trace.RetryConfig{
        Enabled: true,
        // ... 其他重试参数
    }).
    WithOtlpGrpcServiceConfig("...").
    WithOtlpGrpcDialOptions([]grpc.DialOption{}).
    WithOtlpGrpcTLSCredentials(credentials.TransportCredentials).
    WithOtlpGrpcReconnectionPeriod(time.Second * 30).
    WithOtlpGrpcConn(grpcConn)

// OTLP HTTP 专用配置
config.
    WithOtlpHttpPath("/v1/traces").
    WithOtlpHttpTLSClientConfig(&tls.Config{})

// 初始化追踪器
tracer := trace.NewTracer(config)
```

### 2. 追踪器使用

```go
// 启动追踪器
if err := tracer.Start(); err != nil {
    log.Fatal(err)
}
// 关闭追踪器
defer tracer.Stop(context.Background())

// 获取 TracerProvider
tp := tracer.GetTracerProvider()
```

### 3. 与 Gin 框架集成

这里我们使用 otelgin 库来作为 Gin 框架的中间件，更多配置请参考 [otelgin](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/github.com/gin-gonic/gin/otelgin)

```go
// 创建基础配置并初始化 tracer
config := trace.NewTracerConfig().
    WithName("my-service").
    WithProvider(trace.ProviderOtlpGrpc).
    WithEndpoint("localhost:4317")

tracer := trace.NewTracer(config)

// 获取 TracerProvider
tp := tracer.GetTracerProvider()

// 创建 Gin 实例
r := gin.New()

// 配置 Gin 中间件选项
opts := []otelgin.Option{
    // 使用我们的 TracerProvider
    otelgin.WithTracerProvider(tp),
    
    // 设置请求过滤器
    otelgin.WithFilter(func(r *http.Request) bool {
        return !strings.Contains(r.URL.Path, "/health")
    }),
    
    // 设置基于 Gin Context 的过滤器
    otelgin.WithGinFilter(func(c *gin.Context) bool {
        return !strings.HasPrefix(c.FullPath(), "/internal")
    }),
    
    // 自定义 Span 名称格式化
    otelgin.WithSpanNameFormatter(func(r *http.Request) string {
        return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
    }),
}

// 注册中间件，设置服务名称
r.Use(otelgin.Middleware("my-service", opts...))

// 注册路由
r.GET("/api/v1/users", func(c *gin.Context) {
    // 业务逻辑...
    c.JSON(200, gin.H{"message": "success"})
})
```

## 最佳实践

1. **合理设置采样率**

   ```go
   config.WithSamplingRate(0.1) // 采样 10% 的请求
   ```

2. **过滤敏感信息**

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

3. **跳过健康检查等路径**

   ```go
   ginConfig.SkipPaths = []string{
       "/health",
       "/metrics",
       "/favicon.ico",
   }
   ```

4. **使用 Opentelementry Collector 导出数据**

    使用 [opentelemetry-collector](https://github.com/open-telemetry/opentelemetry-collector) 导出数据，作为导出器的中间层，可以更灵活的配置 Sampling 策略和更高效的数据处理，最终导出到 Jaeger 或者 其他后端。

    配置可以参考官方文档 [opentelemetry-collector-configuration](https://opentelemetry.io/docs/collector/configuration/)

## 注意事项

1. Jaeger 导出器已被标记为废弃，建议使用 OTLP 导出器来替代
2. 生产环境建议使用 OTLP 协议连接到可观测性后端
3. 请合理配置采样率，避免产生过多数据
4. 确保正确处理追踪器的启动和关闭生命周期

## 示例代码

完整的示例代码可以参考 `_demo/gin.go`
