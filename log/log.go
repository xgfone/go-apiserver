// Copyright 2021~2022 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package log provides the log function.
package log

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// DefaultLogger is the default logger implementation.
var DefaultLogger Logger

// Pre-define some log levels, which may be assigned to the new values.
var (
	LvlTrace = int(0)
	LvlDebug = int(20)
	LvlInfo  = int(40)
	LvlWarn  = int(60)
	LvlError = int(80)
	LvlAlert = int(100)
	LvlFatal = int(126)
)

// ParseLevel parses the level string, which supports
//   trace
//   debug
//   info
//   warn
//   error
//   alert
//   fatal
// And they are case insensitive.
func ParseLevel(s string) (level int, err error) {
	switch strings.ToLower(s) {
	case "trace":
		level = LvlTrace
	case "debug":
		level = LvlDebug
	case "info":
		level = LvlInfo
	case "warn":
		level = LvlWarn
	case "error":
		level = LvlError
	case "alert":
		level = LvlAlert
	case "fatal":
		level = LvlFatal
	default:
		err = fmt.Errorf("unknown level '%s'", s)
	}
	return
}

// FormatLevel formats the level to string.
func FormatLevel(level int) string {
	switch level {
	case LvlTrace:
		return "trace"
	case LvlDebug:
		return "debug"
	case LvlInfo:
		return "info"
	case LvlWarn:
		return "warn"
	case LvlError:
		return "error"
	case LvlAlert:
		return "alert"
	case LvlFatal:
		return "fatal"
	default:
		if level < LvlDebug {
			return fmt.Sprintf("trace%d", level)
		} else if level < LvlInfo {
			return fmt.Sprintf("debug%d", level)
		} else if level < LvlWarn {
			return fmt.Sprintf("info%d", level)
		} else if level < LvlError {
			return fmt.Sprintf("warn%d", level)
		} else if level < LvlAlert {
			return fmt.Sprintf("error%d", level)
		} else if level < LvlFatal {
			return fmt.Sprintf("alert%d", level)
		} else {
			return fmt.Sprintf("fatal%d", level)
		}
	}
}

// Logger represents a logging implementation.
type Logger interface {
	Enabled(level int) bool
	StdLogger(prefix string, level int) *log.Logger
	Log(level, skipStackDepth int, msg string, keysAndValues ...interface{})
}

// StdLogger returns a stdlib log logger.
func StdLogger(prefix string, level int) *log.Logger {
	return DefaultLogger.StdLogger(prefix, level)
}

// Enabled reports whether the level is enabled.
func Enabled(level int) bool { return DefaultLogger.Enabled(level) }

// Log is used to emit the log.
func Log(level, skipStackDepth int, msg string, keysAndValues ...interface{}) {
	DefaultLogger.Log(level, skipStackDepth+1, msg, keysAndValues...)
}

// Trace is equal to Log(LvlTrace, 0, msg, keysAndValues...).
func Trace(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Log(LvlTrace, 1, msg, keysAndValues...)
}

// Debug is equal to Log(LvlDebug, 0, msg, keysAndValues...).
func Debug(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Log(LvlDebug, 1, msg, keysAndValues...)
}

// Info is equal to Log(LvlInfo, 0, msg, keysAndValues...).
func Info(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Log(LvlInfo, 1, msg, keysAndValues...)
}

// Warn is equal to Log(LvlWarn, 0, msg, keysAndValues...).
func Warn(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Log(LvlWarn, 1, msg, keysAndValues...)
}

// Error is equal to Log(LvlError, 0, msg, keysAndValues...).
func Error(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Log(LvlError, 1, msg, keysAndValues...)
}

// Alert is equal to Log(LvlAlert, 0, msg, keysAndValues...).
func Alert(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Log(LvlAlert, 1, msg, keysAndValues...)
}

// Fatal is equal to Log(LvlFatal, 0, msg, keysAndValues...), then os.Exit(1).
func Fatal(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Log(LvlFatal, 1, msg, keysAndValues...)
	os.Exit(1)
}

func fmtLog(level int, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	DefaultLogger.Log(level, 2, msg)
}

// Tracef is used to emit the log based on the string format with the Trace level.
func Tracef(fmt string, args ...interface{}) { fmtLog(LvlTrace, fmt, args...) }

// Debugf is used to emit the log based on the string format with the Debug level.
func Debugf(fmt string, args ...interface{}) { fmtLog(LvlDebug, fmt, args...) }

// Infof is used to emit the log based on the string format with the Info level.
func Infof(fmt string, args ...interface{}) { fmtLog(LvlInfo, fmt, args...) }

// Warnf is used to emit the log based on the string format with the Warn level.
func Warnf(fmt string, args ...interface{}) { fmtLog(LvlWarn, fmt, args...) }

// Errorf is used to emit the log based on the string format with the Error level.
func Errorf(fmt string, args ...interface{}) { fmtLog(LvlError, fmt, args...) }

// Alertf is used to emit the log based on the string format with the Alert level.
func Alertf(fmt string, args ...interface{}) { fmtLog(LvlAlert, fmt, args...) }

// Fatalf is used to emit the log based on the string format with the Fatal level,
// then exit by os.Exit(1).
func Fatalf(fmt string, args ...interface{}) {
	fmtLog(LvlFatal, fmt, args...)
	os.Exit(1)
}

// Ef is equal to Error(fmt.Sprintf(format, args...), "err", err).
func Ef(err error, format string, args ...interface{}) {
	if len(args) > 0 {
		format = fmt.Sprintf(format, args...)
	}
	DefaultLogger.Log(LvlError, 1, format, "err", err)
}

// IfErr logs the message and key-values with the ERROR level
// only if err is not equal to nil.
func IfErr(err error, msg string, kvs ...interface{}) {
	ifErr(err, 0, msg, kvs...)
}

// WrapPanic wraps and logs the panic, which should be called directly with defer,
// For example,
//   defer WrapPanic()
//   defer WrapPanic("key1", "value1")
//   defer WrapPanic("key1", "value1", "key2", "value2")
//   defer WrapPanic("key1", "value1", "key2", "value2", "key3", "value3")
func WrapPanic(kvs ...interface{}) {
	ifErr(recover(), 1, "panic", kvs...)
}

func ifErr(err interface{}, depth int, msg string, kvs ...interface{}) {
	if err == nil {
		return
	}
	DefaultLogger.Log(LvlError, 2+depth, msg, append(kvs, "err", err)...)
}
