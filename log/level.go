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

package log

import (
	"fmt"
	"strings"
)

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
//
//	trace
//	debug
//	info
//	warn
//	error
//	alert
//	fatal
//
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
