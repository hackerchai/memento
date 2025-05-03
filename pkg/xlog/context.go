package xlog

import (
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog"
)

// LogContext configures a new sub-logger with contextual fields
type LogContext struct {
	logger  *Logger
	context zerolog.Context
}

// Logger returns the logger with the context previously set
func (c LogContext) Logger() *Logger {
	newLogger := *c.logger
	zerologLogger := c.context.Logger()
	newLogger.l = &zerologLogger
	return &newLogger
}

// Fields is a helper function to use a map to set fields using type assertion
func (c LogContext) Fields(fields interface{}) LogContext {
	c.context = c.context.Fields(fields)
	return c
}

// Dict adds the field key with the dict to the logger context
func (c LogContext) Dict(key string, dict *Event) LogContext {
	c.context = c.context.Dict(key, dict)
	return c
}

// Str adds the field key with val as a string to the logger context
func (c LogContext) Str(key string, val string) LogContext {
	c.context = c.context.Str(key, val)
	return c
}

// Strs adds the field key with vals as a []string to the logger context
func (c LogContext) Strs(key string, vals []string) LogContext {
	c.context = c.context.Strs(key, vals)
	return c
}

// Stringer adds the field key with val.String() (or null if val is nil) to the logger context
func (c LogContext) Stringer(key string, val fmt.Stringer) LogContext {
	c.context = c.context.Stringer(key, val)
	return c
}

// Bytes adds the field key with val as a []byte to the logger context
func (c LogContext) Bytes(key string, val []byte) LogContext {
	c.context = c.context.Bytes(key, val)
	return c
}

// Hex adds the field key with val as a hex string to the logger context
func (c LogContext) Hex(key string, val []byte) LogContext {
	c.context = c.context.Hex(key, val)
	return c
}

// RawJSON adds already encoded JSON to context
// No sanity check is performed on b; it must be valid JSON
func (c LogContext) RawJSON(key string, b []byte) LogContext {
	c.context = c.context.RawJSON(key, b)
	return c
}

// AnErr adds the field key with serialized err to the logger context
func (c LogContext) AnErr(key string, err error) LogContext {
	c.context = c.context.AnErr(key, err)
	return c
}

// Err adds the field "error" with serialized err to the logger context
func (c LogContext) Err(err error) LogContext {
	return c.AnErr("error", err)
}

// Bool adds the field key with val as a bool to the logger context
func (c LogContext) Bool(key string, b bool) LogContext {
	c.context = c.context.Bool(key, b)
	return c
}

// Int adds the field key with i as a int to the logger context
func (c LogContext) Int(key string, i int) LogContext {
	c.context = c.context.Int(key, i)
	return c
}

// Float32 adds the field key with f as a float32 to the logger context
func (c LogContext) Float32(key string, f float32) LogContext {
	c.context = c.context.Float32(key, f)
	return c
}

// Float64 adds the field key with f as a float64 to the logger context
func (c LogContext) Float64(key string, f float64) LogContext {
	c.context = c.context.Float64(key, f)
	return c
}

// Time adds the field key with t formatted as string using TimeFieldFormat
func (c LogContext) Time(key string, t time.Time) LogContext {
	c.context = c.context.Time(key, t)
	return c
}

// Dur adds the field key with duration d stored as seconds
func (c LogContext) Dur(key string, d time.Duration) LogContext {
	c.context = c.context.Dur(key, d)
	return c
}

// Interface adds the field key with obj marshaled using reflection
func (c LogContext) Interface(key string, i interface{}) LogContext {
	c.context = c.context.Interface(key, i)
	return c
}

// IPAddr adds IPv4 or IPv6 Address to the context
func (c LogContext) IPAddr(key string, ip net.IP) LogContext {
	c.context = c.context.IPAddr(key, ip)
	return c
}

// IPPrefix adds IPv4 or IPv6 Prefix (address and mask) to the context
func (c LogContext) IPPrefix(key string, pfx net.IPNet) LogContext {
	c.context = c.context.IPPrefix(key, pfx)
	return c
}

// MACAddr adds MAC address to the context
func (c LogContext) MACAddr(key string, ha net.HardwareAddr) LogContext {
	c.context = c.context.MACAddr(key, ha)
	return c
}

// Array adds the field key with an array to the event context
func (c LogContext) Array(key string, arr LogArrayMarshaler) LogContext {
	c.context = c.context.Array(key, arr)
	return c
}

// Object marshals an object that implement the LogObjectMarshaler interface
func (c LogContext) Object(key string, obj LogObjectMarshaler) LogContext {
	c.context = c.context.Object(key, obj)
	return c
}

// EmbedObject marshals and Embeds an object that implement the LogObjectMarshaler interface
func (c LogContext) EmbedObject(obj LogObjectMarshaler) LogContext {
	c.context = c.context.EmbedObject(obj)
	return c
}

// Bools adds the field key with val as a []bool
func (c LogContext) Bools(key string, b []bool) LogContext {
	c.context = c.context.Bools(key, b)
	return c
}

// Ints adds the field key with i as a []int
func (c LogContext) Ints(key string, i []int) LogContext {
	c.context = c.context.Ints(key, i)
	return c
}

// Times adds the field key with t formatted as string using TimeFieldFormat
func (c LogContext) Times(key string, t []time.Time) LogContext {
	c.context = c.context.Times(key, t)
	return c
}

// Timestamp adds the current local time as UNIX timestamp
func (c LogContext) Timestamp() LogContext {
	c.context = c.context.Timestamp()
	return c
}

// Caller adds the file:line of the caller
func (c LogContext) Caller() LogContext {
	c.context = c.context.Caller()
	return c
}

// CallerWithSkipFrameCount adds the file:line of the caller with custom skip frames
func (c LogContext) CallerWithSkipFrameCount(skipFrameCount int) LogContext {
	c.context = c.context.CallerWithSkipFrameCount(skipFrameCount)
	return c
}

// Stack enables stack trace printing for the error
func (c LogContext) Stack() LogContext {
	c.context = c.context.Stack()
	return c
}

// Int8s adds the field key with i as a []int8
func (c LogContext) Int8s(key string, i []int8) LogContext {
	c.context = c.context.Ints8(key, i)
	return c
}

// Int16s adds the field key with i as a []int16
func (c LogContext) Int16s(key string, i []int16) LogContext {
	c.context = c.context.Ints16(key, i)
	return c
}

// Int32s adds the field key with i as a []int32
func (c LogContext) Int32s(key string, i []int32) LogContext {
	c.context = c.context.Ints32(key, i)
	return c
}

// Int64s adds the field key with i as a []int64
func (c LogContext) Int64s(key string, i []int64) LogContext {
	c.context = c.context.Ints64(key, i)
	return c
}

// Uint adds the field key with i as a uint
func (c LogContext) Uint(key string, i uint) LogContext {
	c.context = c.context.Uint(key, i)
	return c
}

// Uints adds the field key with i as a []uint
func (c LogContext) Uints(key string, i []uint) LogContext {
	c.context = c.context.Uints(key, i)
	return c
}

// Uint8 adds the field key with i as a uint8
func (c LogContext) Uint8(key string, i uint8) LogContext {
	c.context = c.context.Uint8(key, i)
	return c
}

// Uint8s adds the field key with i as a []uint8
func (c LogContext) Uint8s(key string, i []uint8) LogContext {
	c.context = c.context.Uints8(key, i)
	return c
}

// Uint16 adds the field key with i as a uint16
func (c LogContext) Uint16(key string, i uint16) LogContext {
	c.context = c.context.Uint16(key, i)
	return c
}

// Uint16s adds the field key with i as a []uint16
func (c LogContext) Uint16s(key string, i []uint16) LogContext {
	c.context = c.context.Uints16(key, i)
	return c
}

// Uint32 adds the field key with i as a uint32
func (c LogContext) Uint32(key string, i uint32) LogContext {
	c.context = c.context.Uint32(key, i)
	return c
}

// Uint32s adds the field key with i as a []uint32
func (c LogContext) Uint32s(key string, i []uint32) LogContext {
	c.context = c.context.Uints32(key, i)
	return c
}

// Uint64 adds the field key with i as a uint64
func (c LogContext) Uint64(key string, i uint64) LogContext {
	c.context = c.context.Uint64(key, i)
	return c
}

// Uint64s adds the field key with i as a []uint64
func (c LogContext) Uint64s(key string, i []uint64) LogContext {
	c.context = c.context.Uints64(key, i)
	return c
}

// Float32s adds the field key with f as a []float32
func (c LogContext) Float32s(key string, f []float32) LogContext {
	c.context = c.context.Floats32(key, f)
	return c
}

// Float64s adds the field key with f as a []float64
func (c LogContext) Float64s(key string, f []float64) LogContext {
	c.context = c.context.Floats64(key, f)
	return c
}
