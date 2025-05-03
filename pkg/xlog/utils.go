package xlog

import (
	"io"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewConsoleWriter creates a new ConsoleWriter with specified configuration
// out: output destination for logs
// timeFormat: format for timestamp in logs
// noColor: whether to disable colored output
func NewConsoleWriter(out io.Writer, timeFormat string, noColor bool) ConsoleWriter {
	return ConsoleWriter{
		Out:        out,
		TimeFormat: timeFormat,
		NoColor:    noColor,
	}
}

// NewMultiWriter creates a multi-writer with multiple writers
func NewMultiWriter(writers ...io.Writer) LevelWriter {
	return zerolog.MultiLevelWriter(writers...)
}

// NewRotateWriter creates a file writer with log rotation support
func NewRotateWriter(opts ...RotateOption) io.Writer {
	// Default configuration
	config := &RotateConfig{
		Filename:   "./app.log", // Log file path
		MaxSize:    100,         // Maximum size of each log file in megabytes
		MaxAge:     7,           // Maximum number of days to retain old log files
		MaxBackups: 10,          // Maximum number of old log files to retain
		Compress:   true,        // Compress/archive old files
		LocalTime:  true,        // Use local time
	}

	// Apply custom options
	for _, opt := range opts {
		opt(config)
	}

	return &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		Compress:   config.Compress,
		LocalTime:  config.LocalTime,
	}
}

// RotateConfig defines configuration for log rotation
type RotateConfig struct {
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
	LocalTime  bool
}

// RotateOption defines the function type for configuration options
type RotateOption func(*RotateConfig)

// WithFilename sets the log file path
func WithFilename(filename string) RotateOption {
	return func(c *RotateConfig) {
		c.Filename = filename
	}
}

// WithMaxSize sets the maximum size in megabytes of a log file before it gets rotated
func WithMaxSize(maxSize int) RotateOption {
	return func(c *RotateConfig) {
		c.MaxSize = maxSize
	}
}

// WithMaxAge sets the maximum number of days to retain old log files
func WithMaxAge(maxAge int) RotateOption {
	return func(c *RotateConfig) {
		c.MaxAge = maxAge
	}
}

// WithMaxBackups sets the maximum number of old log files to retain
func WithMaxBackups(maxBackups int) RotateOption {
	return func(c *RotateConfig) {
		c.MaxBackups = maxBackups
	}
}

// WithCompress enables or disables compression of old log files
func WithCompress(compress bool) RotateOption {
	return func(c *RotateConfig) {
		c.Compress = compress
	}
}

// WithLocalTime sets whether to use local time or UTC for log file names
func WithLocalTime(localTime bool) RotateOption {
	return func(c *RotateConfig) {
		c.LocalTime = localTime
	}
}

// NewBasicSampler creates a basic log sampler
// n: sample every Nth message
func NewBasicSampler(n uint32) zerolog.Sampler {
	return &zerolog.BasicSampler{N: n}
}
