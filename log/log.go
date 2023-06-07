// Copyright 2021~2023 xgfone
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

// Package log provides the log functions based on golang.org/x/exp/slog.
package log

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/io2"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-defaults"
	"golang.org/x/exp/slog"
)

// Predefine some level constants.
const (
	LevelTrace = slog.LevelDebug - 4
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
	LevelFatal = slog.LevelError + 4
)

var (
	// Writer is the default global writer.
	Writer = io2.NewSwitchWriter(os.Stderr)

	// String is a convenient function to new a key-value pair based on string.
	String = slog.String
)

type (
	// Level is the log level.
	Level = slog.Level

	// Leveler is an level interface to get the level.
	Leveler = slog.Leveler

	// LevelVar is used to manage the level atomically, which implements the interface Leveler.
	LevelVar = slog.LevelVar

	// Handler is the log handler.
	Handler = slog.Handler

	// Record is the log record.
	Record = slog.Record

	// Value is the attribute value.
	Value = slog.Value

	// Attr represents a key-value pair.
	Attr = slog.Attr
)

var _ Leveler = LevelFunc(nil)

// LevelFunc is a function to return the log level.
type LevelFunc func() int

// Level implements the interface Leveler to get the log level.
func (f LevelFunc) Level() Level { return Level(f()) }

// NewJSONHandler returns a new json handler.
//
// If w is nil, use Writer instead.
func NewJSONHandler(w io.Writer, level Leveler) Handler {
	if w == nil {
		w = Writer
	}

	return slog.HandlerOptions{
		Level:       level,
		AddSource:   true,
		ReplaceAttr: replaceSourceAttr,
	}.NewJSONHandler(w)
}

func replaceSourceAttr(groups []string, a slog.Attr) slog.Attr {
	switch {
	case a.Key == slog.SourceKey:
		a.Value = slog.StringValue(helper.TrimPkgFile(a.Value.String()))
	case a.Value.Kind() == slog.KindDuration:
		a.Value = slog.StringValue(a.Value.Duration().String())
	}
	return a
}

type OptionHandler struct {
	Handler

	// Options
	EnableFunc  func(context.Context, Level) bool    // Default: nil
	FilterFunc  func(context.Context, Record) bool   // Default: nil
	ReplaceFunc func(context.Context, Record) Record // Default: nil
}

// NewOptionHandler returns a new OptionHandler wrapping the given handler.
func NewOptionHandler(handler Handler) *OptionHandler {
	return &OptionHandler{Handler: handler}
}

func (h *OptionHandler) clone() *OptionHandler {
	nh := *h
	return &nh
}

// Enabled implements the interface Handler#Enabled.
func (h *OptionHandler) Enabled(c context.Context, l Level) bool {
	if h.EnableFunc != nil {
		return h.EnableFunc(c, l)
	}
	return h.Handler.Enabled(c, l)
}

// Handle implements the interface Handler#Handle.
func (h *OptionHandler) Handle(c context.Context, r Record) error {
	if h.ReplaceFunc != nil {
		r = h.ReplaceFunc(c, r)
	}
	if h.FilterFunc != nil && h.FilterFunc(c, r) {
		return nil
	}
	return h.Handler.Handle(c, r)
}

// WithAttrs implements the interface Handler#WithAttrs.
func (h *OptionHandler) WithAttrs(attrs []Attr) Handler {
	nh := h.clone()
	nh.Handler = h.Handler.WithAttrs(attrs)
	return nh
}

// WithGroup implements the interface Handler#WithGroup.
func (h *OptionHandler) WithGroup(name string) Handler {
	nh := h.clone()
	nh.Handler = h.Handler.WithGroup(name)
	return nh
}

// GetHandler returns the handler of the default logger.
func GetHandler() Handler { return slog.Default().Handler() }

// SetDefault is used to set default global logger with the handler.
func SetDefault(handler Handler, atts ...slog.Attr) {
	if len(atts) > 0 {
		handler = handler.WithAttrs(atts)
	}
	log.SetFlags(log.Lshortfile | log.Llongfile)
	slog.SetDefault(slog.New(handler))
}

// Enabled reports whether the level is enabled.
func Enabled(ctx context.Context, level Level) bool {
	return slog.Default().Handler().Enabled(ctx, level)
}

var disableTime bool

// emit is used to emit the log.
func emit(skipStackDepth int, level Level, msg string, kvs ...interface{}) {
	if !Enabled(context.Background(), level) {
		return
	}

	var now time.Time
	if !disableTime {
		now = defaults.Now()
	}

	var pcs [1]uintptr
	runtime.Callers(skipStackDepth+2, pcs[:])
	r := slog.NewRecord(now, level, msg, pcs[0])
	r.Add(kvs...)
	slog.Default().Handler().Handle(context.Background(), r)
}

// Log emits a log with the given level.
func Log(skipStackDepth int, level Level, msg string, kvs ...interface{}) {
	emit(skipStackDepth+1, level, msg, kvs...)
}

// Trace emits a TRACE log.
func Trace(msg string, kvs ...interface{}) { emit(1, LevelTrace, msg, kvs...) }

// Debug emits a DEBUG log.
func Debug(msg string, kvs ...interface{}) { emit(1, LevelDebug, msg, kvs...) }

// Info emits a INFO log.
func Info(msg string, kvs ...interface{}) { emit(1, LevelInfo, msg, kvs...) }

// Warn emits a WARN log.
func Warn(msg string, kvs ...interface{}) { emit(1, LevelWarn, msg, kvs...) }

// Error emits a ERROR log.
func Error(msg string, kvs ...interface{}) { emit(1, LevelError, msg, kvs...) }

// Fatal emits a FATAL log and os.Exit(1).
func Fatal(msg string, kvs ...interface{}) {
	emit(1, LevelFatal, msg, kvs...)
	atexit.Exit(1)
}

func fmtLog(level Level, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	emit(2, level, msg)
}

// Tracef is used to emit the log based on the string format with the TRACE level.
func Tracef(fmt string, args ...interface{}) { fmtLog(LevelTrace, fmt, args...) }

// Debugf is used to emit the log based on the string format with the DEBUG level.
func Debugf(fmt string, args ...interface{}) { fmtLog(LevelDebug, fmt, args...) }

// Infof is used to emit the log based on the string format with the INFO level.
func Infof(fmt string, args ...interface{}) { fmtLog(LevelInfo, fmt, args...) }

// Warnf is used to emit the log based on the string format with the WARN level.
func Warnf(fmt string, args ...interface{}) { fmtLog(LevelWarn, fmt, args...) }

// Errorf is used to emit the log based on the string format with the ERROR level.
func Errorf(fmt string, args ...interface{}) { fmtLog(LevelError, fmt, args...) }

// Fatalf is used to emit the log based on the string format with the FATAL level,
// then exit by os.Exit(1).
func Fatalf(fmt string, args ...interface{}) {
	fmtLog(LevelFatal, fmt, args...)
	atexit.Exit(1)
}

// Ef is equal to Error(fmt.Sprintf(format, args...), "err", err).
func Ef(err error, format string, args ...interface{}) {
	if len(args) > 0 {
		format = fmt.Sprintf(format, args...)
	}
	emit(1, LevelError, format, "err", err)
}

// Err is the same as Error, but appends the error "err" into kvs.
func Err(err error, msg string, kvs ...interface{}) {
	if len(kvs) == 0 {
		kvs = []interface{}{"err", err}
	} else {
		kvs = append(kvs, "err", err)
	}
	emit(1, LevelError, msg, kvs...)
}

// WrapPanic wraps and logs the panic, which should be called directly with defer,
// For example,
//
//	defer WrapPanic()
//	defer WrapPanic("key1", "value1")
//	defer WrapPanic("key1", "value1", "key2", "value2")
//	defer WrapPanic("key1", "value1", "key2", "value2", "key3", "value3")
func WrapPanic(kvs ...interface{}) {
	if r := recover(); r != nil {
		if len(kvs) == 0 {
			kvs = make([]interface{}, 0, 4)
		}

		stacks := helper.GetCallStack(4)
		kvs = append(kvs, "stacks", stacks, "panic", r)
		emit(2, LevelError, "wrap a panic", kvs...)
	}
}
