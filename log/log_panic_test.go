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
	"testing"

	"github.com/xgfone/go-defaults"
)

func TestPanicDefaultHandler(t *testing.T) {
	disableTime = true

	buf := bytes.NewBuffer(nil)
	SetDefault(NewJSONHandler(buf, nil))

	f := func() {
		defer func() {
			if r := recover(); r != nil {
				defaults.HandlePanic(r)
			}
		}()

		panic("test")
	}

	f()

	var info struct {
		Source string   `json:"source"`
		Stacks []string `json:"stacks"`
	}
	if err := json.NewDecoder(buf).Decode(&info); err != nil {
		t.Fatal(err)
	}

	expectSource := "github.com/xgfone/go-apiserver/log/log_panic_test.go:34"
	expectStacksFirst := "github.com/xgfone/go-apiserver/log/log_panic_test.go:func1:38"

	if info.Source != expectSource {
		t.Errorf("expect source '%s', but got '%s'", expectSource, info.Source)
	}
	if len(info.Stacks) == 0 {
		t.Errorf("got 0 stacks")
	} else if info.Stacks[0] != expectStacksFirst {
		t.Errorf("expect stack first line '%s', but got '%s'", expectStacksFirst, info.Stacks[0])
	}
}
