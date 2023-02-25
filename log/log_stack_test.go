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

package log

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

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
	SetDefault(NewJSONHandler(buf, nil))
	testHandleBusiness()

	expects := []string{
		"github.com/xgfone/go-apiserver/log/log_stack_test.go:func1:26",
		"github.com/xgfone/go-apiserver/log/log_stack_test.go:handleBusiness:27",
		"github.com/xgfone/go-apiserver/log/log_stack_test.go:testHandleBusiness:32",
		"github.com/xgfone/go-apiserver/log/log_stack_test.go:TestWrapPanic:38",
	}

	var result struct {
		Stacks []string `json:"stacks"`
	}
	if err := json.NewDecoder(buf).Decode(&result); err != nil {
		t.Fatal(err)
	}

	stacks := result.Stacks
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
