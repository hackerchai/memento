package xlog

import "github.com/rs/zerolog"

type (
	Hook = zerolog.Hook
)

// AddHook add a hook to the logger
func (xlog *Logger) AddHook(hook Hook) *Logger {
	newLogger := xlog.l.Hook(hook)
	xlog.l = &newLogger
	return xlog
}
