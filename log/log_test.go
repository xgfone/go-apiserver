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
	"strings"
	"testing"

	"github.com/xgfone/go-log"
	"github.com/xgfone/go-log/encoder"
)

func newTestEncoder() log.Encoder {
	encoder := encoder.NewJSONEncoder(log.FormatLevel)
	encoder.TimeKey = ""
	return encoder
}

func TestLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	Default().Output.SetEncoder(newTestEncoder())
	Default().SetWriter(buf)
	Info().Kv("k1", "v1").Print("msg1")
	Level(LvlInfo, 0).Kv("k2", "v2").Print("msg2")
	StdLogger("stdlog: ").Print("msg3")

	expects := []string{
		`{"lvl":"info","caller":"log_test.go:36:TestLogger","k1":"v1","msg":"msg1"}`,
		`{"lvl":"info","caller":"log_test.go:37:TestLogger","k2":"v2","msg":"msg2"}`,
		`{"lvl":"debug","caller":"log_test.go:38:TestLogger","msg":"stdlog: msg3"}`,
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
