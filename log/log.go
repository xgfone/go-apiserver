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
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/internal/writer"
	"github.com/xgfone/go-atexit"
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

type (
	// Level is the log level.
	Level = slog.Level

	// Leveler is an level interface to get the level.
	Leveler = slog.Leveler

	// LevelVar is used to manage the level atomically, which implements the interface Leveler.
	LevelVar = slog.LevelVar
)

// NewFileWriter returns a new file writer that rotates the log files
// based on the file size.
//
// filesize is parsed as the log file size, which maybe have a unit suffix,
// such as "123", "123M, 123G". Valid size units contain "b", "B", "k", "K",
// "m", "M", "g", "G", "t", "T", "p", "P", "e", "E". The lower units are 1000x,
// and the upper units are 1024x.
func NewFileWriter(filepath, filesize string, filenum int) (io.WriteCloser, error) {
	if filepath == "" {
		return nil, errors.New("the log filepath must not be empty")
	}

	size, err := writer.ParseSize(filesize)
	if err != nil {
		return nil, err
	}

	return writer.NewSizedRotatingFile(filepath, int(size), filenum), nil
}

// NewJSONHandler returns a new json handler.
func NewJSONHandler(w io.Writer, level Leveler) slog.Handler {
	return slog.HandlerOptions{
		Level:       level,
		AddSource:   true,
		ReplaceAttr: replaceSourceAttr,
	}.NewJSONHandler(w)
}

func replaceSourceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		a.Value = slog.StringValue(helper.TrimPkgFile(a.Value.String()))
	}
	return a
}

// SetDefault is used to set default global logger with the handler.
func SetDefault(ctx context.Context, handler slog.Handler, atts ...slog.Attr) {
	if len(atts) > 0 {
		handler = handler.WithAttrs(atts)
	}

	logger := slog.New(handler)
	if ctx != nil {
		logger = logger.WithContext(ctx)
	}

	log.SetFlags(log.Lshortfile | log.Llongfile)
	slog.SetDefault(logger)
}

// Enabled reports whether the level is enabled.
func Enabled(level Level) bool {
	return slog.Default().Enabled(level)
}

// emit is used to emit the log.
func emit(skipStackDepth int, level Level, msg string, kvs ...interface{}) {
	slog.Default().LogDepth(skipStackDepth+1, level, msg, kvs...)
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
	emit(1, LevelError, format, slog.ErrorKey, err)
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
