// Copyright 2023 xgfone
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

// Package logger provides some log assistants.
package logger

import (
	"context"
	"log/slog"

	"github.com/xgfone/go-defaults"
)

// Trace emits a log message with the TRACE level, which is equal to LevelDebug-4.
func Trace(msg string, args ...any) {
	slog.Log(context.Background(), slog.LevelDebug-4, msg, args...)
}

// Fatal emits a log message with the ERROR level, and call os.Exit(1).
func Fatal(msg string, args ...any) { slog.Error(msg, args...); defaults.Exit(1) }
