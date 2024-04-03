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

package helper

import (
	"errors"
	"fmt"
	"testing"
)

func TestMust(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpect a panic, but got one: %v", r)
			}
		}()
		Must(nil)
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect a panic, but got not")
			} else if s := fmt.Sprint(r); s != "test" {
				t.Errorf("expect '%s', but got '%s'", "test", s)
			}
		}()
		Must(errors.New("test"))
	}()
}
