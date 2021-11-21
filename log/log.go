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

// Package log supplies the log functions.
package log

import (
	stdlog "log"

	"github.com/xgfone/go-log"
)

// Field is the key-value log pair context.
type Field = log.Field

// Define some log functions.
var (
	E         = log.E
	F         = log.F
	Ef        = log.Ef
	IfErr     = log.IfErr
	FieldFunc = log.FieldFunc

	Trace = log.Trace
	Debug = log.Debug
	Info  = log.Info
	Warn  = log.Warn
	Error = log.Error
	Fatal = log.Fatal
)

func init() { log.DefaultLogger.Ctxs = []Field{log.Caller("caller", true)} }

// StdLogger returns a stdlib logger.
func StdLogger(prefix string, flags ...int) *stdlog.Logger {
	return log.DefaultLogger.StdLog(prefix, flags...)
}

// SetNothingWriter sets the writer to the nothing writer to discard all the logs.
func SetNothingWriter() { log.DefaultLogger.Encoder.SetWriter(log.DiscardWriter()) }
