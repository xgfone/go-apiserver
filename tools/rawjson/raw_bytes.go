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

package rawjson

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-pools"
)

// RawBytes represent a raw json bytes.
type RawBytes []byte

// MarshalJSON implements the interface json.Marshaler, which will returns
// the original bytes without the leading and trailing whitespace if not empty,
// else []byte(`""`) instead.
func (s RawBytes) MarshalJSON() ([]byte, error) {
	if js := bytes.TrimSpace(s); len(js) > 0 {
		return js, nil
	}
	return emptyString, nil
}

// WriteTo implements the interface io.WriterTo, which compacts and writes
// the string s as the raw json string into w.
func (s RawBytes) WriteTo(w io.Writer) (n int64, err error) {
	var m int

	if js := bytes.TrimSpace(s); len(js) == 0 {
		m, err = w.Write(emptyString)
	} else {
		buf := pools.GetBuffer(len(js))
		if err = json.Compact(buf.Buffer, js); err == nil {
			m, err = w.Write(buf.Bytes())
		}
		buf.Release()
	}

	n = int64(m)
	return
}

// WriteJSON is the same as WriteTo, but returns nothing.
func (s RawBytes) WriteJSON(w io.Writer) {
	if n, err := s.WriteTo(w); err != nil {
		log.Error("fail to write raw json bytes", "rawjson", string(s), "wrote", n, "err", err)
	}
}
