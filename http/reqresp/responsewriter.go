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
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
)

// GetStatusCode gets the status code from the given value if it implements
// the interface herrors.StatusCodeGetter. Or returns 0.
func GetStatusCode(v interface{}) int {
	if c, ok := v.(StatusCodeGetter); ok {
		return c.StatusCode()
	}
	return 0
}

// ErrIsStatusCode reports whether the error has the given status code,
// which gets the status code from the error by inspect whether it has
// implemented the interface StatusCodeGetter.
func ErrIsStatusCode(err error, statusCode int) bool {
	if c, ok := err.(StatusCodeGetter); ok {
		return statusCode == c.StatusCode()
	}
	return false
}

// StatusCodeGetter is an interface used to get the status code.
type StatusCodeGetter interface {
	StatusCode() int
}

// ResponseWriter is an extended http.ResponseWriter.
type ResponseWriter interface {
	http.ResponseWriter
	StatusCodeGetter
	WroteHeader() bool
	Written() int64
}

// NewResponseWriter returns a new ResponseWriter from http.ResponseWriter
// with the configuration options.
//
// NOTICE: The returned ResponseWriter has also implemented the interface
// { Unwrap() http.ResponseWriter }.
func NewResponseWriter(w http.ResponseWriter, options ...Option) ResponseWriter {
	switch rw := w.(type) {
	case nil:
		return nil
	case ResponseWriter:
		if len(options) == 0 {
			return rw
		}
	}

	rw := &responseWriter{ResponseWriter: w}
	for _, o := range options {
		o.apply(rw)
	}

	var index int
	if _, ok := w.(http.Flusher); ok {
		index += flusher
	}
	if _, ok := w.(http.Hijacker); ok {
		index += hijacker
	}
	if _, ok := w.(io.ReaderFrom); ok {
		index += readerFrom
	}
	if _, ok := w.(http.Pusher); ok {
		index += pusher
	}
	return newResponseWriter(rw, index)
}

// Write returns a ResponseWriter option to wrap the method Write.
func Write(write func([]byte) (int, error)) Option {
	return WriteWithResponse(func(w http.ResponseWriter, b []byte) (int, error) {
		return write(b)
	})
}

// WriteHeader returns a ResponseWriter option to wrap the method WriteHeader.
func WriteHeader(writeHeader func(statusCode int)) Option {
	return WriteHeaderWithResponse(func(rw http.ResponseWriter, statusCode int) {
		writeHeader(statusCode)
	})
}

// WriteWithResponse returns a ResponseWriter option to wrap the method Write.
func WriteWithResponse(write func(rw http.ResponseWriter, data []byte) (int, error)) Option {
	return rwoption(func(w *responseWriter) { w.writeData = write })
}

// WriteHeaderWithResponse returns a ResponseWriter option to wrap the method WriteHeader.
func WriteHeaderWithResponse(writeHeader func(rw http.ResponseWriter, statusCode int)) Option {
	return rwoption(func(w *responseWriter) { w.writeHeader = writeHeader })
}

// Option is used to configure the ResponseWriter.
type Option interface{ apply(*responseWriter) }

type rwoption func(*responseWriter)

func (f rwoption) apply(rw *responseWriter) { f(rw) }

type responseWriter struct {
	http.ResponseWriter

	written     int64
	statusCode  int
	writeHeader func(rw http.ResponseWriter, statusCode int)
	writeData   func(http.ResponseWriter, []byte) (int, error)
}

func (rw *responseWriter) Written() int64    { return rw.written }
func (rw *responseWriter) WroteHeader() bool { return rw.statusCode > 0 }

func (rw *responseWriter) StatusCode() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if statusCode < 100 || statusCode > 999 {
		panic(fmt.Errorf("invalid status code %d", statusCode))
	}

	if rw.statusCode == 0 {
		rw.statusCode = statusCode
		if rw.writeHeader == nil {
			rw.ResponseWriter.WriteHeader(statusCode)
		} else {
			rw.writeHeader(rw.ResponseWriter, statusCode)
		}
	}
}

func (rw *responseWriter) Write(p []byte) (n int, err error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	if rw.writeData == nil {
		n, err = rw.ResponseWriter.Write(p)
	} else {
		n, err = rw.writeData(rw.ResponseWriter, p)
	}
	rw.written += int64(n)
	return
}

func (rw *responseWriter) WriteString(s string) (n int, err error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	if rw.writeData == nil {
		n, err = io.WriteString(rw.ResponseWriter, s)
	} else {
		n, err = rw.writeData(rw.ResponseWriter, []byte(s))
	}
	rw.written += int64(n)
	return
}

func newResponseWriter(rw *responseWriter, index int) ResponseWriter {
	switch index {
	case 0:
		return rw

	case flusher: // 1
		return (*rw1)(rw)

	case hijacker: // 2
		return (*rw2)(rw)

	case hijacker + flusher: // 3
		return (*rw3)(rw)

	case readerFrom: // 4
		return (*rw4)(rw)

	case readerFrom + flusher: // 5
		return (*rw5)(rw)

	case readerFrom + hijacker: // 6
		return (*rw6)(rw)

	case readerFrom + hijacker + flusher: // 7
		return (*rw7)(rw)

	case pusher: // 8
		return (*rw8)(rw)

	case pusher + flusher: // 9
		return (*rw9)(rw)

	case pusher + hijacker: // 10
		return (*rw10)(rw)

	case pusher + hijacker + flusher: // 11
		return (*rw11)(rw)

	case pusher + readerFrom: // 12
		return (*rw12)(rw)

	case pusher + readerFrom + flusher: // 13
		return (*rw13)(rw)

	case pusher + readerFrom + hijacker: // 14
		return (*rw14)(rw)

	case pusher + readerFrom + hijacker + flusher: // 15
		return (*rw15)(rw)

	default:
		panic(fmt.Errorf("unknown ResponseWriter index '%d'", index))
	}
}

/// ----------------------------------------------------------------------- ///

const (
	flusher = 1 << iota
	hijacker
	readerFrom
	pusher
)

func rwFlush(rw *responseWriter) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}
	rw.ResponseWriter.(http.Flusher).Flush()
}

func rwHijack(rw *responseWriter) (net.Conn, *bufio.ReadWriter, error) {
	return rw.ResponseWriter.(http.Hijacker).Hijack()
}

func rwPush(rw *responseWriter, target string, opts *http.PushOptions) error {
	return rw.ResponseWriter.(http.Pusher).Push(target, opts)
}

func rwReadFrom(rw *responseWriter, r io.Reader) (n int64, err error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	n, err = rw.ResponseWriter.(io.ReaderFrom).ReadFrom(r)
	rw.written += n
	return
}

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw1{}
	_ http.Flusher   = &rw1{} // 1
)

type rw1 responseWriter

func (w *rw1) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw1) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw1) Written() int64                    { return w.rw().Written() }
func (w *rw1) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw1) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw1) Header() http.Header               { return w.rw().Header() }
func (w *rw1) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw1) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw1) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw1) Flush() { rwFlush(w.rw()) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw2{}
	_ http.Hijacker  = &rw2{} // 2
)

type rw2 responseWriter

func (w *rw2) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw2) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw2) Written() int64                    { return w.rw().Written() }
func (w *rw2) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw2) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw2) Header() http.Header               { return w.rw().Header() }
func (w *rw2) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw2) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw2) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw2) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw3{}
	_ http.Flusher   = &rw3{} // 1
	_ http.Hijacker  = &rw3{} // 2
)

type rw3 responseWriter

func (w *rw3) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw3) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw3) Written() int64                    { return w.rw().Written() }
func (w *rw3) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw3) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw3) Header() http.Header               { return w.rw().Header() }
func (w *rw3) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw3) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw3) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw3) Flush()                                       { rwFlush(w.rw()) }
func (w *rw3) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw4{}
	_ io.ReaderFrom  = &rw4{} // 4
)

type rw4 responseWriter

func (w *rw4) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw4) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw4) Written() int64                    { return w.rw().Written() }
func (w *rw4) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw4) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw4) Header() http.Header               { return w.rw().Header() }
func (w *rw4) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw4) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw4) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw4) ReadFrom(r io.Reader) (int64, error) { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw5{}
	_ http.Flusher   = &rw5{} // 1
	_ io.ReaderFrom  = &rw5{} // 4
)

type rw5 responseWriter

func (w *rw5) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw5) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw5) Written() int64                    { return w.rw().Written() }
func (w *rw5) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw5) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw5) Header() http.Header               { return w.rw().Header() }
func (w *rw5) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw5) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw5) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw5) Flush()                              { rwFlush(w.rw()) }
func (w *rw5) ReadFrom(r io.Reader) (int64, error) { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw6{}
	_ http.Hijacker  = &rw6{} // 2
	_ io.ReaderFrom  = &rw6{} // 4
)

type rw6 responseWriter

func (w *rw6) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw6) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw6) Written() int64                    { return w.rw().Written() }
func (w *rw6) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw6) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw6) Header() http.Header               { return w.rw().Header() }
func (w *rw6) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw6) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw6) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw6) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }
func (w *rw6) ReadFrom(r io.Reader) (int64, error)          { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw7{}
	_ http.Flusher   = &rw7{} // 1
	_ http.Hijacker  = &rw7{} // 2
	_ io.ReaderFrom  = &rw7{} // 4
)

type rw7 responseWriter

func (w *rw7) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw7) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw7) Written() int64                    { return w.rw().Written() }
func (w *rw7) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw7) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw7) Header() http.Header               { return w.rw().Header() }
func (w *rw7) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw7) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw7) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw7) Flush()                                       { rwFlush(w.rw()) }
func (w *rw7) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }
func (w *rw7) ReadFrom(r io.Reader) (int64, error)          { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw8{}
	_ http.Pusher    = &rw8{} // 8
)

type rw8 responseWriter

func (w *rw8) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw8) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw8) Written() int64                    { return w.rw().Written() }
func (w *rw8) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw8) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw8) Header() http.Header               { return w.rw().Header() }
func (w *rw8) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw8) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw8) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw8) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw9{}
	_ http.Flusher   = &rw9{} // 1
	_ http.Pusher    = &rw9{} // 8
)

type rw9 responseWriter

func (w *rw9) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw9) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw9) Written() int64                    { return w.rw().Written() }
func (w *rw9) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw9) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw9) Header() http.Header               { return w.rw().Header() }
func (w *rw9) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw9) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw9) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw9) Flush()                                           { rwFlush(w.rw()) }
func (w *rw9) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw10{}
	_ http.Hijacker  = &rw10{} // 2
	_ http.Pusher    = &rw10{} // 8
)

type rw10 responseWriter

func (w *rw10) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw10) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw10) Written() int64                    { return w.rw().Written() }
func (w *rw10) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw10) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw10) Header() http.Header               { return w.rw().Header() }
func (w *rw10) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw10) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw10) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw10) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw10) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw11{}
	_ http.Flusher   = &rw11{} // 1
	_ http.Hijacker  = &rw11{} // 2
	_ http.Pusher    = &rw11{} // 8
)

type rw11 responseWriter

func (w *rw11) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw11) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw11) Written() int64                    { return w.rw().Written() }
func (w *rw11) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw11) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw11) Header() http.Header               { return w.rw().Header() }
func (w *rw11) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw11) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw11) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw11) Flush()                                           { rwFlush(w.rw()) }
func (w *rw11) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw11) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw12{}
	_ io.ReaderFrom  = &rw12{} // 4
	_ http.Pusher    = &rw12{} // 8
)

type rw12 responseWriter

func (w *rw12) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw12) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw12) Written() int64                    { return w.rw().Written() }
func (w *rw12) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw12) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw12) Header() http.Header               { return w.rw().Header() }
func (w *rw12) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw12) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw12) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw12) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw12) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw13{}
	_ http.Flusher   = &rw13{} // 1
	_ io.ReaderFrom  = &rw13{} // 4
	_ http.Pusher    = &rw13{} // 8
)

type rw13 responseWriter

func (w *rw13) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw13) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw13) Written() int64                    { return w.rw().Written() }
func (w *rw13) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw13) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw13) Header() http.Header               { return w.rw().Header() }
func (w *rw13) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw13) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw13) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw13) Flush()                                           { rwFlush(w.rw()) }
func (w *rw13) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw13) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw14{}
	_ http.Hijacker  = &rw14{} // 2
	_ io.ReaderFrom  = &rw14{} // 4
	_ http.Pusher    = &rw14{} // 8
)

type rw14 responseWriter

func (w *rw14) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw14) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw14) Written() int64                    { return w.rw().Written() }
func (w *rw14) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw14) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw14) Header() http.Header               { return w.rw().Header() }
func (w *rw14) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw14) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw14) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw14) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw14) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw14) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw15{}
	_ http.Flusher   = &rw15{} // 1
	_ http.Hijacker  = &rw15{} // 2
	_ io.ReaderFrom  = &rw15{} // 4
	_ http.Pusher    = &rw15{} // 8
)

type rw15 responseWriter

func (w *rw15) rw() *responseWriter               { return (*responseWriter)(w) }
func (w *rw15) Unwrap() http.ResponseWriter       { return w.rw().ResponseWriter }
func (w *rw15) Written() int64                    { return w.rw().Written() }
func (w *rw15) StatusCode() int                   { return w.rw().StatusCode() }
func (w *rw15) WroteHeader() bool                 { return w.rw().WroteHeader() }
func (w *rw15) Header() http.Header               { return w.rw().Header() }
func (w *rw15) Write(p []byte) (int, error)       { return w.rw().Write(p) }
func (w *rw15) WriteString(s string) (int, error) { return w.rw().WriteString(s) }
func (w *rw15) WriteHeader(code int)              { w.rw().WriteHeader(code) }

func (w *rw15) Flush()                                           { rwFlush(w.rw()) }
func (w *rw15) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw15) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw15) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }
