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

package logger

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type stacks []string

func (ss stacks) Stacks() []string { return ss }
func (ss stacks) Error() string    { return strings.Join(ss, ", ") }

func TestGetStacks(t *testing.T) {
	expects := []string{"func1", "func2"}
	stacks := getStacks(errors.Join(errors.New("test1"), errors.New("test2"), stacks(expects)))
	if !reflect.DeepEqual(expects, stacks) {
		t.Errorf("expect stacks %v, but got %v", expects, stacks)
	}
}
