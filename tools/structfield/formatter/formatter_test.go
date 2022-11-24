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

package formatter

import (
	"reflect"
	"testing"
)

func TestNameFormatter(t *testing.T) {
	type T struct {
		Name   string
		JSON   string `json:"json"`
		Query  string `query:"query"`
		Header string `header:"header"`
	}

	vt := reflect.TypeOf(T{})
	if field := DefaultNameFormatter(vt.Field(0)); field != "Name" {
		t.Errorf("expect field name '%s', but got '%s'", "Name", field)
	}
	if field := DefaultNameFormatter(vt.Field(1)); field != "json" {
		t.Errorf("expect field name '%s', but got '%s'", "json", field)
	}
	if field := DefaultNameFormatter(vt.Field(2)); field != "query" {
		t.Errorf("expect field name '%s', but got '%s'", "query", field)
	}
	if field := DefaultNameFormatter(vt.Field(3)); field != "header" {
		t.Errorf("expect field name '%s', but got '%s'", "header", field)
	}
}
