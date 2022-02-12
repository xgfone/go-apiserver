// Copyright 2022 xgfone
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

package log

import (
	"fmt"
	"io"
	"log"
	"strings"
)

// NewLogger returns a new Logger with the default implementation.
func NewLogger(out io.Writer, prefix string, flag int) Logger {
	return stdLogger{flag: flag, logger: log.New(out, prefix, flag)}
}

type stdLogger struct {
	flag   int
	logger *log.Logger
}

func (l stdLogger) Enabled(level int) bool { return true }

func (l stdLogger) StdLogger(prefix string, level int) *log.Logger {
	return log.New(l.logger.Writer(), prefix, l.flag)
}

func (l stdLogger) Log(level, depth int, msg string, kvs ...interface{}) {
	var builder strings.Builder
	builder.Grow(128)
	builder.WriteString(msg)
	for i, _len := 0, len(kvs); i < _len; i += 2 {
		fmt.Fprintf(&builder, "; %s=%v", kvs[i], kvs[i+1])
	}
	l.logger.Output(depth+2, builder.String())
}