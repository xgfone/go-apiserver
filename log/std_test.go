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

package log

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestStdLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultLogger = NewLogger(buf, "prefix: ", log.Lshortfile)

	Info("msg1", "k1", "v1")
	Log(LvlInfo, 0, "msg2", "k2", "v2")
	StdLogger("stdlog: ", LvlDebug).Print("msg3")

	expects := []string{
		`prefix: std_test.go:28: msg1; k1=v1`,
		`prefix: std_test.go:29: msg2; k2=v2`,
		`stdlog: std_test.go:30: msg3`,
		``,
	}
	results := strings.Split(buf.String(), "\n")
	if len(expects) != len(results) {
		t.Errorf("expect %d line logs, but got %d", len(expects), len(results))
	} else {
		for i, line := range expects {
			if results[i] != line {
				t.Errorf("%d: expect '%s', but got '%s'", i, line, results[i])
			}
		}
	}
}
