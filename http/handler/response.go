// Copyright 2024 xgfone
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
	"encoding/xml"
	"io"
	"net/http"
	"sync"

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-toolkit/jsonx"
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

	buf := getBuilder()
	if err = jsonx.EncodeJSON(buf, v); err == nil {
		header.SetContentType(w.Header(), header.MIMEApplicationJSONCharsetUTF8)
		w.WriteHeader(code)
		_, err = buf.WriteTo(w)
	}
	putBuilder(buf)

	return
}

// XML sends the response by the xml format to the client.
func XML(w http.ResponseWriter, code int, v any) (err error) {
	if v == nil {
		w.WriteHeader(code)
		return
	}

	buf := getBuilder()
	_, _ = buf.WriteString(xml.Header)
	if err = xml.NewEncoder(buf).Encode(v); err == nil {
		header.SetContentType(w.Header(), header.MIMEApplicationXMLCharsetUTF8)

		w.WriteHeader(code)
		_, err = buf.WriteTo(w)
	}
	putBuilder(buf)

	return
}

/// ----------------------------------------------------------------------- ///

type builder struct{ buf []byte }

func (b *builder) Reset() { b.buf = b.buf[:0] }

func (b *builder) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(b.buf)
	return int64(n), err
}

func (b *builder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *builder) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

var bpool = sync.Pool{New: func() any {
	return &builder{make([]byte, 0, 1024)}
}}

func getBuilder() *builder  { return bpool.Get().(*builder) }
func putBuilder(b *builder) { b.Reset(); bpool.Put(b) }
