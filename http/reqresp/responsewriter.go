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

package reqresp

import (
	"fmt"
	"net/http"
	"sync"
)

// WroteHeader reports whether the response writer has wrote header.
func WroteHeader(w http.ResponseWriter) bool {
	for {
		switch rw := w.(type) {
		case interface{ WroteHeader() bool }:
			return rw.WroteHeader()

		case interface{ Unwrap() http.ResponseWriter }:
			w = rw.Unwrap()

		default:
			return false
		}
	}
}

// ResponseWriter is an extended http.ResponseWriter.
type ResponseWriter interface {
	http.ResponseWriter
	WroteHeader() bool
	StatusCode() int
}

// AcquireResponseWriter acquires a ResponseWriter with w from the pool.
func AcquireResponseWriter(w http.ResponseWriter) ResponseWriter {
	rw := rwpool.Get().(*responseWriter)
	rw.Reset(w)
	return rw
}

// ReleaseResponseWriter releases the ResponseWriter into the pool.
func ReleaseResponseWriter(w ResponseWriter) {
	if rw, ok := w.(*responseWriter); ok {
		rw.Reset(nil)
		rwpool.Put(rw)
	}
}

var rwpool = &sync.Pool{New: func() any { return newResponseWriter() }}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter() *responseWriter { return new(responseWriter) }

func (r *responseWriter) Unwrap() http.ResponseWriter { return r.ResponseWriter }

func (r *responseWriter) StatusCode() int {
	if r.statusCode == 0 {
		return 200
	}
	return r.statusCode
}

func (r *responseWriter) WroteHeader() bool {
	return r.statusCode > 0
}

func (r *responseWriter) WriteHeader(code int) {
	if code < 100 {
		panic(fmt.Errorf("invalid http response status code %d", code))
	}

	if r.statusCode == 0 {
		r.statusCode = code
		r.ResponseWriter.WriteHeader(code)
	}
}

func (r *responseWriter) Reset(w http.ResponseWriter) {
	*r = responseWriter{ResponseWriter: w, statusCode: 0}
}
