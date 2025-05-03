//nolint:reassign
package xlog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type XLogReqIDKey string

const (
	// DefaultCallerDepth is the default caller depth
	// Because the caller of the logger is wrapped by the xlog, so the default depth is 3
	DefaultCallerDepth = 3
	DefaultLogLevel    = zerolog.InfoLevel
	DefaultReqIDKey    = XLogReqIDKey("X-Reqid")
)

const (
	Ltrace   = zerolog.TraceLevel
	Ldebug   = zerolog.DebugLevel
	Linfo    = zerolog.InfoLevel
	Lwarn    = zerolog.WarnLevel
	Lerror   = zerolog.ErrorLevel
	Lfatal   = zerolog.FatalLevel
	Lpanic   = zerolog.PanicLevel
	LnoLevel = zerolog.NoLevel
)

const (
	TimeFormatUnix      = zerolog.TimeFormatUnix
	TimeFormatUnixMs    = zerolog.TimeFormatUnixMs
	TimeFormatUnixMicro = zerolog.TimeFormatUnixMicro
	TimeFormatUnixNano  = zerolog.TimeFormatUnixNano
)

type (
	Level              = zerolog.Level
	Context            = zerolog.Context
	Event              = zerolog.Event
	LogObjectMarshaler = zerolog.LogObjectMarshaler
	LogArrayMarshaler  = zerolog.LogArrayMarshaler
	ConsoleWriter      = zerolog.ConsoleWriter
	LevelWriter        = zerolog.LevelWriter
	Sampler            = zerolog.Sampler
	BasicSampler       = zerolog.BasicSampler
	LevelSampler       = zerolog.LevelSampler
	BurstSampler       = zerolog.BurstSampler
)

// ctxKey is the key for the logger in the context
type ctxKey struct{}

type Logger struct {
	calldepth int
	l         *zerolog.Logger
	// trace is whether to add trace information to the log
	trace bool
	// reqIDKey is the key for the request ID in the context
	reqIDKey string
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() *Logger {
	zerologLogger := zerolog.New(os.Stdout).Level(DefaultLogLevel).With().Timestamp().CallerWithSkipFrameCount(DefaultCallerDepth).Logger()

	return &Logger{
		calldepth: DefaultCallerDepth,
		l:         &zerologLogger,
		trace:     true,
		reqIDKey:  string(DefaultReqIDKey),
	}
}

// NewLogger creates a new logger with the given writer, level, and caller depth
// w is the log output writer
// level is the log level
// timeFormat is the time format
// timeLocation is the time location
func NewLogger(w io.Writer, level Level, timeFormat string, timeLocation *time.Location) *Logger {
	zerologLogger := zerolog.New(w).Level(level).With().Timestamp().CallerWithSkipFrameCount(DefaultCallerDepth).Logger()
	if timeFormat != "" {
		zerolog.TimeFieldFormat = timeFormat
	}
	if timeLocation != nil {
		zerolog.TimestampFunc = func() time.Time {
			return time.Now().In(timeLocation)
		}
	}

	return &Logger{
		calldepth: DefaultCallerDepth,
		l:         &zerologLogger,
		trace:     true,
		reqIDKey:  string(DefaultReqIDKey),
	}
}

// WithContext put logger into context
func (xlog *Logger) WithContext(ctx context.Context) context.Context {
	if _, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, xlog)
}

// Ctx get logger from context
func Ctx(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return l
	} else {
		return NewDefaultLogger()
	}
}

// Sample sets the sampler for the logger
func (xlog *Logger) Sample(sampler Sampler) *Logger {
	logger := xlog.l.Sample(sampler)
	xlog.l = &logger
	return xlog
}

// SetLevel sets the log level
func (xlog *Logger) SetLevel(level Level) *Logger {
	logger := xlog.l.Level(level)
	xlog.l = &logger
	return xlog
}

// SetCaller sets the caller information
// depth: call stack depth, if negative, it will use the zerolog global caller depth
func (xlog *Logger) SetCaller(depth int) *Logger {
	xlog.calldepth = depth
	logger := xlog.l.With().CallerWithSkipFrameCount(depth).Logger()
	xlog.l = &logger
	return xlog
}

// SetTimeField sets the time field name
func (xlog *Logger) SetTimeField(field string) *Logger {
	zerolog.TimestampFieldName = field
	return xlog
}

// SetTimeFormat sets the time format
func (xlog *Logger) SetTimeFormat(format string) *Logger {
	zerolog.TimeFieldFormat = format
	return xlog
}

// SetTimeLocation sets the time location
func (xlog *Logger) SetTimeLocation(loc *time.Location) *Logger {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(loc)
	}
	return xlog
}

// SetLevelField sets the level field name
func (xlog *Logger) SetLevelField(field string) *Logger {
	zerolog.LevelFieldName = field
	return xlog
}

// SetMessageField sets the message field name
func (xlog *Logger) SetMessageField(field string) *Logger {
	zerolog.MessageFieldName = field
	return xlog
}

// SetErrorField sets the error field name
func (xlog *Logger) SetErrorField(field string) *Logger {
	zerolog.ErrorFieldName = field
	return xlog
}

// SetCallerField sets the caller field name
func (xlog *Logger) SetCallerField(field string) *Logger {
	zerolog.CallerFieldName = field
	return xlog
}

// SetLevelTraceValue sets the trace level value
func (xlog *Logger) SetLevelTraceValue(value string) *Logger {
	zerolog.LevelTraceValue = value
	return xlog
}

// SetLevelDebugValue sets the debug level value
func (xlog *Logger) SetLevelDebugValue(value string) *Logger {
	zerolog.LevelDebugValue = value
	return xlog
}

// SetLevelInfoValue sets the info level value
func (xlog *Logger) SetLevelInfoValue(value string) *Logger {
	zerolog.LevelInfoValue = value
	return xlog
}

// SetLevelWarnValue sets the warn level value
func (xlog *Logger) SetLevelWarnValue(value string) *Logger {
	zerolog.LevelWarnValue = value
	return xlog
}

// SetLevelErrorValue sets the error level value
func (xlog *Logger) SetLevelErrorValue(value string) *Logger {
	zerolog.LevelErrorValue = value
	return xlog
}

// SetLevelFatalValue sets the fatal level value
func (xlog *Logger) SetLevelFatalValue(value string) *Logger {
	zerolog.LevelFatalValue = value
	return xlog
}

// SetLevelPanicValue sets the panic level value
func (xlog *Logger) SetLevelPanicValue(value string) *Logger {
	zerolog.LevelPanicValue = value
	return xlog
}

// SetErrorStackFieldName sets the error stack field name
func (xlog *Logger) SetErrorStackFieldName(field string) *Logger {
	zerolog.ErrorStackFieldName = field
	return xlog
}

// SetErrorHandler sets the error handler
func (xlog *Logger) SetErrorHandler(fn func(err error)) *Logger {
	zerolog.ErrorHandler = fn
	return xlog
}

// SetErrorStackMarshalFunc sets the error stack marshal function
func (xlog *Logger) SetErrorStackMarshalFunc(fn func(err error) interface{}) *Logger {
	zerolog.ErrorStackMarshaler = fn
	return xlog
}

// SetErrorMarshalFunc sets the error marshal function
func (xlog *Logger) SetErrorMarshalFunc(fn func(err error) interface{}) *Logger {
	zerolog.ErrorMarshalFunc = fn
	return xlog
}

// SetInterfaceMarshalFunc sets the interface marshal function
func (xlog *Logger) SetInterfaceMarshalFunc(fn func(v any) ([]byte, error)) *Logger {
	zerolog.InterfaceMarshalFunc = fn
	return xlog
}

// SetLevelFieldMarshalFunc sets the level field marshal function
func (xlog *Logger) SetLevelFieldMarshalFunc(fn func(l Level) string) *Logger {
	zerolog.LevelFieldMarshalFunc = fn
	return xlog
}

// SetCallerMarshalFunc sets the caller marshal function
func (xlog *Logger) SetCallerMarshalFunc(fn func(pc uintptr, file string, line int) string) *Logger {
	zerolog.CallerMarshalFunc = fn
	return xlog
}

// GetCallerDepth returns the caller depth
func (xlog *Logger) GetCallerDepth() int {
	return xlog.calldepth
}

// GetTimeField returns the time field name
func (xlog *Logger) GetTimeField() string {
	return zerolog.TimestampFieldName
}

// GetTimeFormat returns the time format
func (xlog *Logger) GetTimeFormat() string {
	return zerolog.TimeFieldFormat
}

// GetTimeLocation returns the time location
func (xlog *Logger) GetTimeLocation() *time.Location {
	return zerolog.TimestampFunc().Location()
}

// GetLevelField returns the level field name
func (xlog *Logger) GetLevelField() string {
	return zerolog.LevelFieldName
}

// GetMessageField returns the message field name
func (xlog *Logger) GetMessageField() string {
	return zerolog.MessageFieldName
}

// GetErrorField returns the error field name
func (xlog *Logger) GetErrorField() string {
	return zerolog.ErrorFieldName
}

// GetCallerField returns the caller field name
func (xlog *Logger) GetCallerField() string {
	return zerolog.CallerFieldName
}

// GetLevelTraceValue returns the trace level value
func (xlog *Logger) GetLevelTraceValue() string {
	return zerolog.LevelTraceValue
}

// GetLevelDebugValue returns the debug level value
func (xlog *Logger) GetLevelDebugValue() string {
	return zerolog.LevelDebugValue
}

// GetLevelInfoValue returns the info level value
func (xlog *Logger) GetLevelInfoValue() string {
	return zerolog.LevelInfoValue
}

// GetLevelWarnValue returns the warn level value
func (xlog *Logger) GetLevelWarnValue() string {
	return zerolog.LevelWarnValue
}

// GetLevelErrorValue returns the error level value
func (xlog *Logger) GetLevelErrorValue() string {
	return zerolog.LevelErrorValue
}

// GetLevelFatalValue returns the fatal level value
func (xlog *Logger) GetLevelFatalValue() string {
	return zerolog.LevelFatalValue
}

// GetLevelPanicValue returns the panic level value
func (xlog *Logger) GetLevelPanicValue() string {
	return zerolog.LevelPanicValue
}

// GetErrorStackFieldName returns the error stack field name
func (xlog *Logger) GetErrorStackFieldName() string {
	return zerolog.ErrorStackFieldName
}

// GetErrorHandler returns the error handler
func (xlog *Logger) GetErrorHandler() func(err error) {
	return zerolog.ErrorHandler
}

// GetErrorStackMarshaler returns the error stack marshal function
func (xlog *Logger) GetErrorStackMarshaler() func(err error) interface{} {
	return zerolog.ErrorStackMarshaler
}

// GetErrorMarshalFunc returns the error marshal function
func (xlog *Logger) GetErrorMarshalFunc() func(err error) interface{} {
	return zerolog.ErrorMarshalFunc
}

// GetInterfaceMarshalFunc returns the interface marshal function
func (xlog *Logger) GetInterfaceMarshalFunc() func(v any) ([]byte, error) {
	return zerolog.InterfaceMarshalFunc
}

// GetLevelFieldMarshalFunc returns the level field marshal function
func (xlog *Logger) GetLevelFieldMarshalFunc() func(l Level) string {
	return zerolog.LevelFieldMarshalFunc
}

// GetCallerMarshalFunc returns the caller marshal function
func (xlog *Logger) GetCallerMarshalFunc() func(pc uintptr, file string, line int) string {
	return zerolog.CallerMarshalFunc
}

func (xlog *Logger) GetLevel() Level {
	return xlog.l.GetLevel()
}

// SetTrace sets the trace information
func (xlog *Logger) SetTrace(trace bool) *Logger {
	xlog.trace = trace
	return xlog
}

// SetReqIDKey sets the request ID key
func (xlog *Logger) SetReqIDKey(key string) *Logger {
	xlog.reqIDKey = key
	return xlog
}

// ReqIDFromContext returns the request ID from the context
func (xlog *Logger) ReqIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(XLogReqIDKey(xlog.reqIDKey)).(string); ok {
		return reqID
	}
	return ""
}

// SetReqIDToContext sets the request ID to the context
func (xlog *Logger) SetReqIDToContext(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, XLogReqIDKey(xlog.reqIDKey), reqID)
}

// NewContext creates a new context for the logger
func (xlog *Logger) With() LogContext {
	return LogContext{
		logger:  xlog,
		context: xlog.l.With(),
	}
}

// UpdateContext updates the logger context value
// Be careful when calling UpdateContext. It is not concurrency safe. Use the With method to create a child logger
func (xlog *Logger) UpdateContext(fn func(c Context) Context) *Logger {
	xlog.l.UpdateContext(fn)
	return xlog
}

// SetWriter sets the log output writer
func (xlog *Logger) SetWriter(w io.Writer) *Logger {
	logger := xlog.l.Output(w)
	xlog.l = &logger
	return xlog
}

// Build constructs and returns the configured Logger
func (xlog *Logger) Build() *Logger {
	return xlog
}

// Clone creates a copy of the current Logger efficiently
func (xlog *Logger) Clone() *Logger {
	if xlog == nil {
		return nil
	}

	newLogger := *xlog

	// Shallow copy internal logger
	if xlog.l != nil {
		newInternalLogger := *xlog.l
		newLogger.l = &newInternalLogger
	}

	newLogger.trace = xlog.trace
	newLogger.reqIDKey = xlog.reqIDKey

	return &newLogger
}

// TraceX returns a new Event with trace level
// if ctx is provided and trace is enabled, it will add trace information to the event
func (xlog *Logger) TraceX(ctx ...context.Context) *LogEvent {
	if len(ctx) > 0 && xlog.trace {
		return xlog.newEventWithTrace(Ltrace, xlog.l.Trace(), ctx[0])
	}

	return &LogEvent{logger: xlog, event: xlog.l.Trace(), level: Ltrace}
}

// DebugX returns a new Event with debug level
// if ctx is provided and trace is enabled, it will add trace information to the event
func (xlog *Logger) DebugX(ctx ...context.Context) *LogEvent {
	if len(ctx) > 0 && xlog.trace {
		return xlog.newEventWithTrace(Ldebug, xlog.l.Debug(), ctx[0])
	}

	return &LogEvent{logger: xlog, event: xlog.l.Debug(), level: Ldebug}
}

// InfoX returns a new Event with info level
// if ctx is provided and trace is enabled, it will add trace information to the event
func (xlog *Logger) InfoX(ctx ...context.Context) *LogEvent {
	if len(ctx) > 0 && xlog.trace {
		return xlog.newEventWithTrace(Linfo, xlog.l.Info(), ctx[0])
	}

	return &LogEvent{logger: xlog, event: xlog.l.Info(), level: Linfo}
}

// WarnX returns a new Event with warn level
// if ctx is provided and trace is enabled, it will add trace information to the event
func (xlog *Logger) WarnX(ctx ...context.Context) *LogEvent {
	if len(ctx) > 0 && xlog.trace {
		return xlog.newEventWithTrace(Lwarn, xlog.l.Warn(), ctx[0])
	}

	return &LogEvent{logger: xlog, event: xlog.l.Warn(), level: Lwarn}
}

// ErrorX returns a new Event with error level
// if ctx is provided and trace is enabled, it will add trace information to the event
func (xlog *Logger) ErrorX(ctx ...context.Context) *LogEvent {
	if len(ctx) > 0 && xlog.trace {
		return xlog.newEventWithTrace(Lerror, xlog.l.Error(), ctx[0])
	}

	return &LogEvent{logger: xlog, event: xlog.l.Error(), level: Lerror}
}

// FatalX returns a new Event with fatal level
// if ctx is provided and trace is enabled, it will add trace information to the event
func (xlog *Logger) FatalX(ctx ...context.Context) *LogEvent {
	if len(ctx) > 0 && xlog.trace {
		return xlog.newEventWithTrace(Lfatal, xlog.l.Fatal(), ctx[0])
	}

	return &LogEvent{logger: xlog, event: xlog.l.Fatal(), level: Lfatal}
}

// PanicX returns a new Event with panic level
// if ctx is provided and trace is enabled, it will add trace information to the event
func (xlog *Logger) PanicX(ctx ...context.Context) *LogEvent {
	if len(ctx) > 0 && xlog.trace {
		return xlog.newEventWithTrace(Lpanic, xlog.l.Panic(), ctx[0])
	}

	return &LogEvent{logger: xlog, event: xlog.l.Panic(), level: Lpanic}
}

// TraceMsg logs a message with trace level
func (xlog *Logger) TraceMsg(msg string) {
	xlog.l.Trace().Msg(msg)
}

// TraceMsgf logs a formatted message with trace level
func (xlog *Logger) TraceMsgf(format string, v ...interface{}) {
	xlog.l.Trace().Msgf(format, v...)
}

// DebugMsg logs a message with debug level
func (xlog *Logger) DebugMsg(msg string) {
	xlog.l.Debug().Msg(msg)
}

// DebugMsgf logs a formatted message with debug level
func (xlog *Logger) DebugMsgf(format string, v ...interface{}) {
	xlog.l.Debug().Msgf(format, v...)
}

// InfoMsg logs a message with info level
func (xlog *Logger) InfoMsg(msg string) {
	event := xlog.InfoX()
	event.Msg(msg)
}

// InfoMsgf logs a formatted message with info level
func (xlog *Logger) InfoMsgf(format string, v ...interface{}) {
	event := xlog.InfoX()
	event.Msgf(format, v...)
}

// WarnMsg logs a message with warn level
func (xlog *Logger) WarnMsg(msg string) {
	event := xlog.WarnX()
	event.Msg(msg)
}

// WarnMsgf logs a formatted message with warn level
func (xlog *Logger) WarnMsgf(format string, v ...interface{}) {
	event := xlog.WarnX()
	event.Msgf(format, v...)
}

// ErrorMsg logs a message with error level
func (xlog *Logger) ErrorMsg(msg string) {
	event := xlog.ErrorX()
	event.Msg(msg)
}

// ErrorMsgf logs a formatted message with error level
func (xlog *Logger) ErrorMsgf(format string, v ...interface{}) {
	event := xlog.ErrorX()
	event.Msgf(format, v...)
}

// FatalMsg logs a message with fatal level
func (xlog *Logger) FatalMsg(msg string) {
	event := xlog.FatalX()
	event.Msg(msg)
}

// FatalMsgf logs a formatted message with fatal level
func (xlog *Logger) FatalMsgf(format string, v ...interface{}) {
	event := xlog.FatalX()
	event.Msgf(format, v...)
}

// PanicMsg logs a message with panic level
func (xlog *Logger) PanicMsg(msg string) {
	event := xlog.PanicX()
	event.Msg(msg)
}

// PanicMsgf logs a formatted message with panic level
func (xlog *Logger) PanicMsgf(format string, v ...interface{}) {
	event := xlog.PanicX()
	event.Msgf(format, v...)
}

// From creates a new Logger from a zerolog.Logger
func From(zerologLogger *zerolog.Logger) *Logger {
	return &Logger{
		l:        zerologLogger,
		trace:    true,
		reqIDKey: string(DefaultReqIDKey),
	}
}

// GetLogger returns the underlying zerolog.Logger
func (xlog *Logger) GetLogger() (*zerolog.Logger, error) {
	if xlog.l == nil {
		return nil, errors.New("logger is not initialized")
	}
	return xlog.l, nil
}

// Unwrap returns the underlying zerolog.Logger
func (xlog *Logger) Unwrap() *zerolog.Logger {
	return xlog.l
}

// Print calls Output to print to logger
func (xlog *Logger) Print(v ...interface{}) {
	xlog.l.Print(v...)
}

// Printf calls Output to print to logger with formatting
func (xlog *Logger) Printf(format string, v ...interface{}) {
	xlog.l.Printf(format, v...)
}

// Println calls Output to print to logger with a newline
func (xlog *Logger) Println(v ...interface{}) {
	xlog.l.Println(v...)
}

// Debugf logs a formatted debug message with request ID
func (xlog *Logger) Debugf(format string, v ...interface{}) {
	xlog.l.Debug().Msgf(format, v...)
}

// Debug logs a debug message with request ID
func (xlog *Logger) Debug(v ...interface{}) {
	xlog.l.Debug().Msg(fmt.Sprint(v...))
}

// Infof logs a formatted info message with request ID
func (xlog *Logger) Infof(format string, v ...interface{}) {
	xlog.l.Info().Msgf(format, v...)
}

// Info logs an info message with request ID
func (xlog *Logger) Info(v ...interface{}) {
	xlog.l.Info().Msg(fmt.Sprint(v...))
}

// Warnf logs a formatted warning message with request ID
func (xlog *Logger) Warnf(format string, v ...interface{}) {
	xlog.l.Warn().Msgf(format, v...)
}

// Warn logs a warning message with request ID
func (xlog *Logger) Warn(v ...interface{}) {
	xlog.l.Warn().Msg(fmt.Sprint(v...))
}

// Errorf logs a formatted error message with request ID
func (xlog *Logger) Errorf(format string, v ...interface{}) {
	xlog.l.Error().Msgf(format, v...)
}

// Error logs an error message with request ID
func (xlog *Logger) Error(v ...interface{}) {
	xlog.l.Error().Msg(fmt.Sprint(v...))
}

// Panicf logs a formatted panic message with request ID and then panics
func (xlog *Logger) Panicf(format string, v ...interface{}) {
	xlog.l.Panic().Msgf(format, v...)
}

// Panic logs a panic message with request ID and then panics
func (xlog *Logger) Panic(v ...interface{}) {
	xlog.l.Panic().Msg(fmt.Sprint(v...))
}

// Panicln logs a panic message with newline and request ID, then panics
func (xlog *Logger) Panicln(v ...interface{}) {
	xlog.l.Panic().Msg(fmt.Sprintln(v...))
}

// Fatalf logs a formatted fatal message with request ID and then exits
func (xlog *Logger) Fatalf(format string, v ...interface{}) {
	xlog.l.Fatal().Msgf(format, v...)
}

// Fatal logs a fatal message with request ID and then exits
func (xlog *Logger) Fatal(v ...interface{}) {
	xlog.l.Fatal().Msg(fmt.Sprint(v...))
}

// Fatalln logs a fatal message with newline and request ID, then exits
func (xlog *Logger) Fatalln(v ...interface{}) {
	xlog.l.Fatal().Msg(fmt.Sprintln(v...))
}

// Stack logs an error message with full stack trace
func (xlog *Logger) Stack(v ...interface{}) {
	s := generateStackString(v, true)
	xlog.l.Error().Msg(s)
}

// SingleStack logs an error message with single goroutine stack trace
func (xlog *Logger) SingleStack(v ...interface{}) {
	s := generateStackString(v, false)

	xlog.l.Error().Msg(s)
}

// generateStackString generates a stack trace string from the given arguments
// Parameters:
//   - v: arguments to format into message
//   - full: whether to include full goroutine stack trace
func generateStackString(v []interface{}, full bool) string {
	var sb strings.Builder
	msg := fmt.Sprint(v...)
	sb.WriteString(msg)
	sb.WriteString("\n")

	buf := make([]byte, 1024*1024)
	n := runtime.Stack(buf, full)
	stack := string(buf[:n])

	sb.WriteString(stack)
	return sb.String()
}
