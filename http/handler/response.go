// Copyright 2024~2025 xgfone
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

package handler

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/internal/pools"
	"github.com/xgfone/go-toolkit/httpx"
	"github.com/xgfone/go-toolkit/jsonx"
	"github.com/xgfone/go-toolkit/unsafex"
)

// JSONResponder is an interface to send the http response to client.
type JSONResponder interface {
	JSON(code int, value any)
}

// JSON sends the response by the json format to the client.
func JSON(w http.ResponseWriter, code int, v any) (err error) {
	if v == nil {
		w.WriteHeader(code)
		return
	}

	pool, buf := pools.GetBuffer(64 * 1024) // 64KB
	defer pools.PutBuffer(pool, buf)

	if err = jsonx.Marshal(buf, v); err == nil {
		header.SetContentType(w.Header(), httpx.MIMEApplicationJSONCharsetUTF8)
		w.WriteHeader(code)
		err = write(w, buf)
	}

	return
}

// XML sends the response by the xml format to the client.
func XML(w http.ResponseWriter, code int, v any) (err error) {
	if v == nil {
		w.WriteHeader(code)
		return
	}

	pool, buf := pools.GetBuffer(64 * 1024) // 64KB
	defer pools.PutBuffer(pool, buf)

	_, _ = buf.WriteString(xml.Header)
	if err = xml.NewEncoder(buf).Encode(v); err == nil {
		header.SetContentType(w.Header(), httpx.MIMEApplicationXMLCharsetUTF8)
		w.WriteHeader(code)
		err = write(w, buf)
	}

	return
}

func write(w http.ResponseWriter, b *bytes.Buffer) (err error) {
	n, err := w.Write(unsafex.Bytes(b.String()))
	if err == nil && n != b.Len() {
		err = io.ErrShortWrite
	}
	return
}
