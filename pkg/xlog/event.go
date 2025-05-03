package xlog

import (
	"fmt"
	"net"
	"time"
)

type LogEvent struct {
	logger *Logger
	event  *Event
	level  Level
}

func (xlog *Logger) newEvent(level Level, event *Event) *LogEvent {
	e := &LogEvent{
		logger: xlog,
		event:  event,
		level:  level,
	}

	return e
}

// Msg logs a message
func (e *LogEvent) Msg(msg string) {
	e.event.Msg(msg)
}

// Msgf logs a formatted message
func (e *LogEvent) Msgf(format string, args ...interface{}) {
	e.event.Msgf(format, args...)
}

// Int adds an integer field to the event
func (e *LogEvent) Int(key string, value int) *LogEvent {
	e.event.Int(key, value)
	return e
}

// Str adds a string field to the event
func (e *LogEvent) Str(key string, value string) *LogEvent {
	e.event.Str(key, value)
	return e
}

func (e *LogEvent) Strs(key string, value []string) *LogEvent {
	e.event.Strs(key, value)
	return e
}

// Stringer adds a stringer field to the event
func (e *LogEvent) Stringer(key string, value fmt.Stringer) *LogEvent {
	e.event.Stringer(key, value)
	return e
}

// Stringers adds a slice of stringers field to the event
func (e *LogEvent) Stringers(key string, value []fmt.Stringer) *LogEvent {
	e.event.Stringers(key, value)
	return e
}

func (e *LogEvent) Any(key string, value interface{}) *LogEvent {
	e.event.Any(key, value)
	return e
}

// Bool adds a boolean field to the event
func (e *LogEvent) Bool(key string, value bool) *LogEvent {
	e.event.Bool(key, value)
	return e
}

// Bools adds a slice of booleans field to the event
func (e *LogEvent) Bools(key string, value []bool) *LogEvent {
	e.event.Bools(key, value)
	return e
}

// Dur adds a duration field to the event
func (e *LogEvent) Dur(key string, value time.Duration) *LogEvent {
	e.event.Dur(key, value)
	return e
}

// Durs adds a slice of durations field to the event
func (e *LogEvent) Durs(key string, value []time.Duration) *LogEvent {
	e.event.Durs(key, value)
	return e
}

// Err adds an error field to the event
func (e *LogEvent) Err(err error) *LogEvent {
	e.event.Err(err)
	return e
}

// Errs adds a slice of errors field to the event
func (e *LogEvent) Errs(key string, errs []error) *LogEvent {
	e.event.Errs(key, errs)
	return e
}

// Float64 adds a float64 field to the event
func (e *LogEvent) Float64(key string, value float64) *LogEvent {
	e.event.Float64(key, value)
	return e
}

// Floats64 adds a slice of float64s field to the event
func (e *LogEvent) Floats64(key string, value []float64) *LogEvent {
	e.event.Floats64(key, value)
	return e
}

// Uint64 adds a uint64 field to the event
func (e *LogEvent) Uint64(key string, value uint64) *LogEvent {
	e.event.Uint64(key, value)
	return e
}

// Uint32 adds a uint32 field to the event
func (e *LogEvent) Uint32(key string, value uint32) *LogEvent {
	e.event.Uint32(key, value)
	return e
}

// Uint16 adds a uint16 field to the event
func (e *LogEvent) Uint16(key string, value uint16) *LogEvent {
	e.event.Uint16(key, value)
	return e
}

// Uint8 adds a uint8 field to the event
func (e *LogEvent) Uint8(key string, value uint8) *LogEvent {
	e.event.Uint8(key, value)
	return e
}

// RawJSON adds a json.RawMessage field to the event
func (e *LogEvent) RawJSON(key string, value []byte) *LogEvent {
	e.event.RawJSON(key, value)
	return e
}

// Bytes adds a byte slice field to the event
func (e *LogEvent) Bytes(key string, value []byte) *LogEvent {
	e.event.Bytes(key, value)
	return e
}

// Hex adds a byte slice as hex string field to the event
func (e *LogEvent) Hex(key string, value []byte) *LogEvent {
	e.event.Hex(key, value)
	return e
}

// IPAddr adds an IP address field to the event
func (e *LogEvent) IPAddr(key string, value net.IP) *LogEvent {
	e.event.IPAddr(key, value)
	return e
}

// IPPrefix adds an IP prefix field to the event
func (e *LogEvent) IPPrefix(key string, value net.IPNet) *LogEvent {
	e.event.IPPrefix(key, value)
	return e
}

// MACAddr adds a MAC address field to the event
func (e *LogEvent) MACAddr(key string, value net.HardwareAddr) *LogEvent {
	e.event.MACAddr(key, value)
	return e
}

// Type adds a field with the type of the value
func (e *LogEvent) Type(key string, v interface{}) *LogEvent {
	e.event.Type(key, v)
	return e
}

// Caller adds caller information to the event
func (e *LogEvent) Caller(depth int) *LogEvent {
	e.event.Caller(depth)
	return e
}

// Stack adds stack trace information to the event
func (e *LogEvent) Stack() *LogEvent {
	e.event.Stack()
	return e
}

// Enabled returns whether the event is enabled
func (e *LogEvent) Enabled() bool {
	return e.event.Enabled()
}

// Discard discards the event
func (e *LogEvent) Discard() *LogEvent {
	e.event = e.event.Discard()
	return e
}

// Time adds a time.Time field to the event
func (e *LogEvent) Time(key string, t time.Time) *LogEvent {
	e.event.Time(key, t)
	return e
}

// Times adds a slice of time.Time field to the event
func (e *LogEvent) Times(key string, a []time.Time) *LogEvent {
	e.event.Times(key, a)
	return e
}

// TimeDiff adds the time difference between two time.Time to the event
func (e *LogEvent) TimeDiff(key string, t time.Time, start time.Time) *LogEvent {
	e.event.TimeDiff(key, t, start)
	return e
}

// Float32 adds a float32 field to the event
func (e *LogEvent) Float32(key string, f float32) *LogEvent {
	e.event.Float32(key, f)
	return e
}

// Floats32 adds a slice of float32s field to the event
func (e *LogEvent) Floats32(key string, f []float32) *LogEvent {
	e.event.Floats32(key, f)
	return e
}

// Ints64 adds a slice of int64s field to the event
func (e *LogEvent) Ints64(key string, a []int64) *LogEvent {
	e.event.Ints64(key, a)
	return e
}

// Ints32 adds a slice of int32s field to the event
func (e *LogEvent) Ints32(key string, a []int32) *LogEvent {
	e.event.Ints32(key, a)
	return e
}

// Ints16 adds a slice of int16s field to the event
func (e *LogEvent) Ints16(key string, a []int16) *LogEvent {
	e.event.Ints16(key, a)
	return e
}

// Ints8 adds a slice of int8s field to the event
func (e *LogEvent) Ints8(key string, a []int8) *LogEvent {
	e.event.Ints8(key, a)
	return e
}

// Ints adds a slice of ints field to the event
func (e *LogEvent) Ints(key string, a []int) *LogEvent {
	e.event.Ints(key, a)
	return e
}

// Uints64 adds a slice of uint64s field to the event
func (e *LogEvent) Uints64(key string, a []uint64) *LogEvent {
	e.event.Uints64(key, a)
	return e
}

// Uints32 adds a slice of uint32s field to the event
func (e *LogEvent) Uints32(key string, a []uint32) *LogEvent {
	e.event.Uints32(key, a)
	return e
}

// Uints16 adds a slice of uint16s field to the event
func (e *LogEvent) Uints16(key string, a []uint16) *LogEvent {
	e.event.Uints16(key, a)
	return e
}

// Uints8 adds a slice of uint8s field to the event
func (e *LogEvent) Uints8(key string, a []uint8) *LogEvent {
	e.event.Uints8(key, a)
	return e
}

// Uints adds a slice of uints field to the event
func (e *LogEvent) Uints(key string, a []uint) *LogEvent {
	e.event.Uints(key, a)
	return e
}

// Interface adds an interface{} field to the event
func (e *LogEvent) Interface(key string, i interface{}) *LogEvent {
	e.event.Interface(key, i)
	return e
}

// Object adds an object that implements log.ObjectMarshaler to the event
func (e *LogEvent) Object(key string, obj LogObjectMarshaler) *LogEvent {
	e.event.Object(key, obj)
	return e
}

// EmbedObject embeds an object that implements log.ObjectMarshaler to the event
func (e *LogEvent) EmbedObject(obj LogObjectMarshaler) *LogEvent {
	e.event.EmbedObject(obj)
	return e
}

// Func executes a function only if the event is enabled
func (e *LogEvent) Func(f func(*LogEvent)) *LogEvent {
	if f != nil {
		f(e)
	}
	return e
}

// Dict adds a dictionary to the event
func (e *LogEvent) Dict(key string, dict *Event) *LogEvent {
	e.event.Dict(key, dict)
	return e
}

// Fields adds multiple fields to the event
// Only map[string]interface{} and []interface{} are accepted. []interface{} must
// alternate string keys and arbitrary values, and extraneous ones are ignored.
func (e *LogEvent) Fields(fields interface{}) *LogEvent {
	e.event.Fields(fields)
	return e
}

// Int8 adds an int8 field to the event
func (e *LogEvent) Int8(key string, value int8) *LogEvent {
	if e == nil {
		return e
	}
	e.event.Int8(key, value)
	return e
}

// Int16 adds an int16 field to the event
func (e *LogEvent) Int16(key string, value int16) *LogEvent {
	if e == nil {
		return e
	}
	e.event.Int16(key, value)
	return e
}

// Int32 adds an int32 field to the event
func (e *LogEvent) Int32(key string, value int32) *LogEvent {
	if e == nil {
		return e
	}
	e.event.Int32(key, value)
	return e
}

// Int64 adds an int64 field to the event
func (e *LogEvent) Int64(key string, value int64) *LogEvent {
	if e == nil {
		return e
	}
	e.event.Int64(key, value)
	return e
}

// Timestamp adds the current time as UNIX timestamp to the event
func (e *LogEvent) Timestamp() *LogEvent {
	if e == nil {
		return e
	}
	e.event.Timestamp()
	return e
}

// Array adds an array to the event
func (e *LogEvent) Array(key string, arr LogArrayMarshaler) *LogEvent {
	if e == nil {
		return e
	}
	e.event.Array(key, arr)
	return e
}

// Send is equivalent to calling Msg("")
func (e *LogEvent) Send() {
	if e == nil {
		return
	}
	e.event.Send()
}

// MsgFunc sends the event with a message computed by the function
func (e *LogEvent) MsgFunc(fn func() string) {
	if e == nil {
		return
	}
	e.event.MsgFunc(fn)
}

// CallerSkipFrame sets the number of stack frames to skip
func (e *LogEvent) CallerSkipFrame(skip int) *LogEvent {
	if e == nil {
		return e
	}
	e.event.CallerSkipFrame(skip)
	return e
}

// AnErr is an alias for Err
func (e *LogEvent) AnErr(key string, err error) *LogEvent {
	if e == nil {
		return e
	}
	e.event.AnErr(key, err)
	return e
}

// Uint adds an unsigned integer field to the event
func (e *LogEvent) Uint(key string, v uint) *LogEvent {
	if e == nil {
		return e
	}
	e.event.Uint(key, v)
	return e
}

func (e *LogEvent) Logger() *Logger {
	return e.logger
}
