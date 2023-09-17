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
	"bytes"
	"strings"
	"testing"
)

func TestDecodeJSONFromBytes(t *testing.T) {
	var v int
	if err := DecodeJSONFromBytes(&v, []byte(`123`)); err != nil {
		t.Error(err)
	} else if v != 123 {
		t.Errorf("expect %d, but got %d", 123, v)
	}
}

func TestDecodeJSONFromString(t *testing.T) {
	var v int
	if err := DecodeJSONFromString(&v, `123`); err != nil {
		t.Error(err)
	} else if v != 123 {
		t.Errorf("expect %d, but got %d", 123, v)
	}
}

func TestEncodeJSON(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	if err := EncodeJSON(buf, `abc`); err != nil {
		t.Error(err)
	} else if s := strings.TrimSpace(buf.String()); s != `"abc"` {
		t.Errorf("expect '%s', but got '%s'", `"abc"`, s)
	}
}

func TestEncodeJSONBytes(t *testing.T) {
	if data, err := EncodeJSONBytes(`abc`); err != nil {
		t.Error(err)
	} else if s := string(data); s != `"abc"` {
		t.Errorf("expect '%s', but got '%s'", `"abc"`, s)
	}
}

func TestEncodeJSONString(t *testing.T) {
	if s, err := EncodeJSONString(`abc`); err != nil {
		t.Error(err)
	} else if s != `"abc"` {
		t.Errorf("expect '%s', but got '%s'", `"abc"`, s)
	}
}
