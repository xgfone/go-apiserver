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

// Package log provides the log functions, which is deprecated and reserved
// in order to be compatible.
package log

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/xgfone/go-apiserver/io2"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-defaults"
)

// Define some extra levels.
const (
	LevelTrace = slog.LevelDebug - 4
	LevelFatal = slog.LevelError + 4
)

// Writer is the default global writer.
var Writer = io2.NewSwitchWriter(os.Stderr)

// NewJSONHandler returns a new json handler.
//
// If w is nil, use Writer instead.
func NewJSONHandler(w io.Writer, level slog.Leveler) slog.Handler {
	if w == nil {
		w = Writer
	}

	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       level,
		AddSource:   true,
		ReplaceAttr: replaceSourceAttr,
	})
}

func replaceSourceAttr(groups []string, a slog.Attr) slog.Attr {
	switch {
	case a.Key == slog.SourceKey:
		if src, ok := a.Value.Any().(*slog.Source); ok {
			a.Value = slog.StringValue(fmt.Sprintf("%s:%d", defaults.TrimPkgFile(src.File), src.Line))
		}
	case a.Key == slog.LevelKey:
		if lvl, ok := a.Value.Any().(slog.Level); ok {
			switch lvl {
			case LevelTrace:
				a.Value = slog.StringValue("TRACE")
			case LevelFatal:
				a.Value = slog.StringValue("FATAL")
			}
		}
	case a.Value.Kind() == slog.KindDuration:
		a.Value = slog.StringValue(a.Value.Duration().String())
	}
	return a
}

type OptionHandler struct {
	slog.Handler

	// Options
	EnableFunc  func(context.Context, slog.Level) bool         // Default: nil
	FilterFunc  func(context.Context, slog.Record) bool        // Default: nil
	ReplaceFunc func(context.Context, slog.Record) slog.Record // Default: nil
}

// NewOptionHandler returns a new OptionHandler wrapping the given handler.
func NewOptionHandler(handler slog.Handler) *OptionHandler {
	return &OptionHandler{Handler: handler}
}

func (h *OptionHandler) clone() *OptionHandler {
	nh := *h
	return &nh
}

// Enabled implements the interface Handler#Enabled.
func (h *OptionHandler) Enabled(c context.Context, l slog.Level) bool {
	if h.EnableFunc != nil {
		return h.EnableFunc(c, l)
	}
	return h.Handler.Enabled(c, l)
}

// Handle implements the interface Handler#Handle.
func (h *OptionHandler) Handle(c context.Context, r slog.Record) error {
	if h.ReplaceFunc != nil {
		r = h.ReplaceFunc(c, r)
	}
	if h.FilterFunc != nil && h.FilterFunc(c, r) {
		return nil
	}
	return h.Handler.Handle(c, r)
}

// WithAttrs implements the interface Handler#WithAttrs.
func (h *OptionHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := h.clone()
	nh.Handler = h.Handler.WithAttrs(attrs)
	return nh
}

// WithGroup implements the interface Handler#WithGroup.
func (h *OptionHandler) WithGroup(name string) slog.Handler {
	nh := h.clone()
	nh.Handler = h.Handler.WithGroup(name)
	return nh
}

// GetHandler returns the handler of the default logger.
func GetHandler() slog.Handler { return slog.Default().Handler() }

// SetDefault is used to set default global logger with the handler.
func SetDefault(handler slog.Handler, atts ...slog.Attr) {
	if len(atts) > 0 {
		handler = handler.WithAttrs(atts)
	}
	log.SetFlags(log.Lshortfile | log.Llongfile)
	slog.SetDefault(slog.New(handler))
}

// Enabled reports whether the level is enabled.
func Enabled(ctx context.Context, level slog.Level) bool {
	return slog.Default().Handler().Enabled(ctx, level)
}

var disableTime bool

// emit is used to emit the log.
func emit(skipStackDepth int, level slog.Level, msg string, kvs ...any) {
	if !Enabled(context.Background(), level) {
		return
	}

	var now time.Time
	if !disableTime {
		now = time.Now()
	}

	var pcs [1]uintptr
	runtime.Callers(skipStackDepth+2, pcs[:])
	r := slog.NewRecord(now, level, msg, pcs[0])
	r.Add(kvs...)
	_ = slog.Default().Handler().Handle(context.Background(), r)
}

// Log emits a log with the given level.
func Log(skipStackDepth int, level slog.Level, msg string, kvs ...any) {
	emit(skipStackDepth+1, level, msg, kvs...)
}

// Trace emits a TRACE log.
func Trace(msg string, kvs ...any) { emit(1, LevelTrace, msg, kvs...) }

// Debug emits a DEBUG log.
func Debug(msg string, kvs ...any) { emit(1, slog.LevelDebug, msg, kvs...) }

// Info emits a INFO log.
func Info(msg string, kvs ...any) { emit(1, slog.LevelInfo, msg, kvs...) }

// Warn emits a WARN log.
func Warn(msg string, kvs ...any) { emit(1, slog.LevelWarn, msg, kvs...) }

// Error emits a ERROR log.
func Error(msg string, kvs ...any) { emit(1, slog.LevelError, msg, kvs...) }

// Fatal emits a FATAL log and os.Exit(1).
func Fatal(msg string, kvs ...any) {
	emit(1, LevelFatal, msg, kvs...)
	atexit.Exit(1)
}

func fmtLog(level slog.Level, msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	emit(2, level, msg)
}

// Tracef is used to emit the log based on the string format with the TRACE level.
func Tracef(fmt string, args ...any) { fmtLog(LevelTrace, fmt, args...) }

// Debugf is used to emit the log based on the string format with the DEBUG level.
func Debugf(fmt string, args ...any) { fmtLog(slog.LevelDebug, fmt, args...) }

// Infof is used to emit the log based on the string format with the INFO level.
func Infof(fmt string, args ...any) { fmtLog(slog.LevelInfo, fmt, args...) }

// Warnf is used to emit the log based on the string format with the WARN level.
func Warnf(fmt string, args ...any) { fmtLog(slog.LevelWarn, fmt, args...) }

// Errorf is used to emit the log based on the string format with the ERROR level.
func Errorf(fmt string, args ...any) { fmtLog(slog.LevelError, fmt, args...) }

// Fatalf is used to emit the log based on the string format with the FATAL level,
// then exit by os.Exit(1).
func Fatalf(fmt string, args ...any) {
	fmtLog(LevelFatal, fmt, args...)
	atexit.Exit(1)
}

// WrapPanic wraps and logs the panic, which should be called directly with defer.
func WrapPanic() {
	if r := recover(); r != nil {
		logpanic(r, 3)
	}
}

func init() { defaults.HandlePanicFunc.Set(func(r any) { logpanic(r, 5) }) }

func logpanic(r any, skip int) {
	stacks := defaults.GetStacks(skip)
	emit(3, slog.LevelError, "wrap a panic", slog.Any("panic", r), slog.Any("stacks", stacks))
}
