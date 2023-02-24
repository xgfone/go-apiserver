// Copyright 2022~2023 xgfone
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
	"log"
	"runtime"
	"time"

	"golang.org/x/exp/slog"
)

// StdLogger returns a stdlib log logger.
func StdLogger(prefix string, level slog.Level) *log.Logger {
	writer := &handlerWriter{handler: slog.Default().Handler(), level: level}
	return log.New(writer, prefix, 0)
}

type handlerWriter struct {
	handler slog.Handler
	level   slog.Level
}

func (w *handlerWriter) Write(buf []byte) (int, error) {
	if !w.handler.Enabled(nil, w.level) {
		return 0, nil
	}

	// Remove final newline.
	origLen := len(buf)
	if len(buf) > 0 && buf[len(buf)-1] == '\n' {
		buf = buf[:len(buf)-1]
	}

	r := slog.NewRecord(time.Now(), w.level, string(buf), callerPC(5))
	return origLen, w.handler.Handle(nil, r)
}

// callerPC returns the program counter at the given stack depth.
func callerPC(depth int) uintptr {
	var pcs [1]uintptr
	runtime.Callers(depth, pcs[:])
	return pcs[0]
}
