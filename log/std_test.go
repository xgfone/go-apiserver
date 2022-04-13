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
	"errors"
	"log"
	"strings"
	"testing"
)

func TestStdLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultLogger = NewLogger(buf, "prefix: ", log.Lshortfile, LvlTrace)

	Info("msg1", "k1", "v1")
	Log(LvlInfo, 0, "msg2", "k2", "v2")
	StdLogger("stdlog: ", LvlDebug).Print("msg3")
	Infof("msg4: %s=%s", "k3", "v3")
	IfErr(errors.New("error"), "msg5")

	expects := []string{
		`prefix: std_test.go:29: msg1; level=info; k1=v1`,
		`prefix: std_test.go:30: msg2; level=info; k2=v2`,
		`stdlog: std_test.go:31: msg3`,
		`prefix: std_test.go:32: msg4: k3=v3; level=info`,
		`prefix: std_test.go:33: msg5; level=error; err=error`,
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

func handleBusiness() {
	func(s string) {
		panic(s)
	}("test")
}

func testHandleBusiness() {
	defer WrapPanic()
	handleBusiness()
}

func TestWrapPanic(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultLogger = NewLogger(buf, "", log.Lshortfile, LvlTrace)

	testHandleBusiness()

	results := strings.Split(buf.String(), "\n")
	if len(results) != 2 {
		t.Errorf("expect %d line logs, but got %d", 2, len(results))
		return
	}

	prefix := "std_test.go:57: wrap a panic; level=error; stacks=["
	if !strings.HasPrefix(results[0], prefix) {
		t.Errorf("unexpected line: %s", results[0])
		return
	}

	stack := results[0][len(prefix):]
	if index := strings.IndexByte(stack, ']'); index > -1 {
		stack = stack[:index]
	}

	expects := []string{
		"github.com/xgfone/go-apiserver/log/std_test.go:func1:57",
		"github.com/xgfone/go-apiserver/log/std_test.go:handleBusiness:58",
		"github.com/xgfone/go-apiserver/log/std_test.go:testHandleBusiness:63",
		"github.com/xgfone/go-apiserver/log/std_test.go:TestWrapPanic:70",
	}

	stacks := strings.Fields(stack)
	for i, stack := range stacks { // Remove the testing.go.
		if strings.HasPrefix(stack, "testing/testing.go:") {
			stacks = stacks[:i]
			break
		}
	}

	if len(expects) != len(stacks) {
		t.Errorf("expect %d stacks, but got %d", len(expects), len(stacks))
	} else {
		for i, line := range expects {
			if stacks[i] != line {
				t.Errorf("%d: expect stack '%s', but got '%s'", i, line, stacks[i])
			}
		}
	}

}
