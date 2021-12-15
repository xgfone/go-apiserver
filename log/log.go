// Copyright 2021 xgfone
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

// Package log provides the log functions.
package log

import (
	stdlog "log"

	"github.com/xgfone/go-log"
	"github.com/xgfone/go-log/logf"
)

// Pre-define some log levels.
const (
	LvlTrace   = log.LvlTrace
	LvlDebug   = log.LvlDebug
	LvlInfo    = log.LvlInfo
	LvlWarn    = log.LvlWarn
	LvlError   = log.LvlError
	LvlAlert   = log.LvlAlert
	LvlPanic   = log.LvlPanic
	LvlFatal   = log.LvlFatal
	LvlDisable = log.LvlDisable
)

// Pre-define some log functions.
var (
	Trace func() log.Logger = log.Trace
	Debug func() log.Logger = log.Debug
	Info  func() log.Logger = log.Info
	Warn  func() log.Logger = log.Warn
	Error func() log.Logger = log.Error
	Alert func() log.Logger = log.Alert
	Panic func() log.Logger = log.Panic
	Fatal func() log.Logger = log.Fatal

	Tracef func(msg string, args ...interface{}) = logf.Tracef
	Debugf func(msg string, args ...interface{}) = logf.Debugf
	Infof  func(msg string, args ...interface{}) = logf.Infof
	Warnf  func(msg string, args ...interface{}) = logf.Warnf
	Errorf func(msg string, args ...interface{}) = logf.Errorf
	Alertf func(msg string, args ...interface{}) = logf.Alertf
	Panicf func(msg string, args ...interface{}) = logf.Panicf
	Fatalf func(msg string, args ...interface{}) = logf.Fatalf

	Kv     func(key string, value interface{}) log.Logger = log.Kv
	Kvs    func(kvs ...interface{}) log.Logger            = log.Kvs
	Print  func(args ...interface{})                      = log.Print
	Printf func(msg string, args ...interface{})          = log.Printf

	Ef        func(err error, msg string, args ...interface{}) = log.Ef
	IfErr     func(err error, msg string, kvs ...interface{})  = log.IfErr
	WrapPanic func(kvs ...interface{})                         = log.WrapPanic
)

type jsonEncoder = log.JSONEncoder

// Default returns the default logger.
func Default() *log.Engine { return log.DefaultLogger }

// Clone clones the default logger and returns a new one.
func Clone() *log.Engine { return log.DefaultLogger.Clone() }

// Logger returns a log emitter based on the level.
func Logger(level, depth int) log.Logger { return log.Log(level, depth+1) }

// StdLogger returns a stdlib log logger.
func StdLogger(prefix string) *stdlog.Logger { return log.StdLog(prefix) }
