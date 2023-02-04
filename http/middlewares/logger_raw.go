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

package middlewares

import (
	"encoding"
	"encoding/json"
	"strings"

	"github.com/xgfone/go-apiserver/internal/pools"
)

var emptyString = []byte(`""`)

// rawString represent a raw json string.
type rawString string

var (
	_ json.Marshaler         = rawString("")
	_ encoding.TextMarshaler = rawString("")
)

// MarshalJSON implements the interface json.Marshaler, which will returns
// the original string without the leading and trailing whitespace as []byte
// if not empty, else []byte(`""`) instead.
func (s rawString) MarshalJSON() ([]byte, error) {
	if js := strings.TrimSpace(string(s)); len(js) > 0 {
		return []byte(js), nil
	}
	return emptyString, nil
}

func (s rawString) MarshalText() ([]byte, error) {
	js, err := s.compact()
	return []byte(js), err
}

func (s rawString) compact() (js string, err error) {
	if js = strings.TrimSpace(string(s)); len(js) > 0 {
		pool, buf := pools.GetBuffer(len(js))
		if err = json.Compact(buf, []byte(js)); err == nil {
			js = buf.String()
		}
		pools.PutBuffer(pool, buf)
	}
	return
}
