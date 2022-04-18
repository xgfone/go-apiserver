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

package helper

import (
	"fmt"
	"runtime"
	"strings"
)

var trimPrefixes = []string{"/src/", "/pkg/mod/"}

// RecoverStackSkip is used to skip some stacks in recover.
const RecoverStackSkip = 4

// GetCallStack returns the most 64 call stacks.
func GetCallStack(skip int) []string {
	var pcs [64]uintptr
	n := runtime.Callers(skip, pcs[:])
	if n == 0 {
		return nil
	}

	stacks := make([]string, 0, n)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		for _, mark := range trimPrefixes {
			if index := strings.Index(frame.File, mark); index > -1 {
				frame.File = frame.File[index+len(mark):]
				break
			}
		}

		if frame.Function == "" {
			stacks = append(stacks, fmt.Sprintf("%s:%d", frame.File, frame.Line))
		} else {
			name := frame.Function
			if index := strings.LastIndexByte(frame.Function, '.'); index > -1 {
				name = frame.Function[index+1:]
			}
			stacks = append(stacks, fmt.Sprintf("%s:%s:%d", frame.File, name, frame.Line))
		}
	}

	return stacks
}
