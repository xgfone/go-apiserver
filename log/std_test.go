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
	"encoding/json"
	"testing"

	"golang.org/x/exp/slog"
)

func TestStdLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	SetDefault(nil, NewJSONHandler(buf, nil), slog.String("ctxkey", "ctxvalue"))

	// Emit the log by the std log.Logger
	StdLogger("stdlog: ", slog.LevelError).Printf("msg")

	type Result struct {
		Level  string `json:"level"`
		Source string `json:"source"`
		Msg    string `json:"msg"`
		CtxKey string `json:"ctxkey"`
	}

	expect := Result{
		Level:  "ERROR",
		Source: "github.com/xgfone/go-apiserver/log/std_test.go:30",
		Msg:    "stdlog: msg",
		CtxKey: "ctxvalue",
	}

	var result Result
	if err := json.NewDecoder(buf).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result != expect {
		t.Errorf("expect %+v, but got %+v", expect, result)
	}
}
