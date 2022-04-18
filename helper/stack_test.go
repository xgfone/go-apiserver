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
	"strings"
	"testing"
)

func TestGetCallStack(t *testing.T) {
	stacks := GetCallStack(1)
	for i, stack := range stacks {
		if strings.HasPrefix(stack, "testing/") {
			stacks = stacks[:i]
			break
		}
	}

	expects := []string{
		"github.com/xgfone/go-apiserver/helper/stack.go:GetCallStack:31",
		"github.com/xgfone/go-apiserver/helper/stack_test.go:TestGetCallStack:23",
	}

	if len(expects) != len(stacks) {
		t.Fatalf("expect %d line, but got %d: %v", len(expects), len(stacks), stacks)
	}

	for i, line := range expects {
		if line != stacks[i] {
			t.Errorf("%d: expect '%s', but got '%s'", i, line, stacks[i])
		}
	}
}
