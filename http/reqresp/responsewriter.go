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

// Unwrap unwraps the innest wrapped http ResponseWriter.
//
// If rw has not implemented the interface WrappedResponseWriter, return itself.
func Unwrap(rw http.ResponseWriter) http.ResponseWriter {
	for {
		ww, ok := rw.(WrappedResponseWriter)
		if !ok {
			return rw
		}
		rw = ww.WrappedResponseWriter()
	}
}

// WrappedResponseWriter is used to unwrap the wrapped http.ResponseWriter.
type WrappedResponseWriter interface {
	WrappedResponseWriter() http.ResponseWriter
	http.ResponseWriter
}

// ResponseWriter is an extended http.ResponseWriter.
type ResponseWriter interface {
	http.ResponseWriter
	WroteHeader() bool
	StatusCode() int
	Written() int64
}

type responseWriter struct {
	http.ResponseWriter

	written    int64
	statusCode int
}

func (rw *responseWriter) Written() int64    { return rw.written }
func (rw *responseWriter) WroteHeader() bool { return rw.statusCode > 0 }

func (rw *responseWriter) StatusCode() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}

func (rw *responseWriter) WrappedResponseWriter() http.ResponseWriter {
	return rw.ResponseWriter
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if statusCode < 100 || statusCode > 999 {
		panic(fmt.Errorf("invalid status code %d", statusCode))
	}

	if rw.statusCode == 0 {
		rw.statusCode = statusCode
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *responseWriter) Write(p []byte) (n int, err error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	n, err = rw.ResponseWriter.Write(p)
	rw.written += int64(n)
	return
}

func (rw *responseWriter) WriteString(s string) (n int, err error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	n, err = io.WriteString(rw.ResponseWriter, s)
	rw.written += int64(n)
	return
}

// NewResponseWriter returns a new ResponseWriter from http.ResponseWriter.
func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	if w == nil {
		panic("http.ResponseWriter is nil")
	}

	rw := &responseWriter{ResponseWriter: w}

	var index int
	if _, ok := w.(http.CloseNotifier); ok {
		index += closeNotifier
	}
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

func newResponseWriter(rw *responseWriter, index int) ResponseWriter {
	switch index {
	case 0:
		return rw

	case closeNotifier: // 1
		return (*rw1)(rw)

	case flusher: // 2
		return (*rw2)(rw)

	case flusher + closeNotifier: // 3
		return (*rw3)(rw)

	case hijacker: // 4
		return (*rw4)(rw)

	case hijacker + closeNotifier: // 5
		return (*rw5)(rw)

	case hijacker + flusher: // 6
		return (*rw6)(rw)

	case hijacker + flusher + closeNotifier: // 7
		return (*rw7)(rw)

	case readerFrom: // 8
		return (*rw8)(rw)

	case readerFrom + closeNotifier: // 9
		return (*rw9)(rw)

	case readerFrom + flusher: // 10
		return (*rw10)(rw)

	case readerFrom + flusher + closeNotifier: // 11
		return (*rw11)(rw)

	case readerFrom + hijacker: // 12
		return (*rw12)(rw)

	case readerFrom + hijacker + closeNotifier: // 13
		return (*rw13)(rw)

	case readerFrom + hijacker + flusher: // 14
		return (*rw14)(rw)

	case readerFrom + hijacker + flusher + closeNotifier: // 15
		return (*rw15)(rw)

	case pusher: // 16
		return (*rw16)(rw)

	case pusher + closeNotifier: // 17
		return (*rw17)(rw)

	case pusher + flusher: // 18
		return (*rw18)(rw)

	case pusher + flusher + closeNotifier: // 19
		return (*rw19)(rw)

	case pusher + hijacker: // 20
		return (*rw20)(rw)

	case pusher + hijacker + closeNotifier: // 21
		return (*rw21)(rw)

	case pusher + hijacker + flusher: // 22
		return (*rw22)(rw)

	case pusher + hijacker + flusher + closeNotifier: //23
		return (*rw23)(rw)

	case pusher + readerFrom: // 24
		return (*rw24)(rw)

	case pusher + readerFrom + closeNotifier: // 25
		return (*rw25)(rw)

	case pusher + readerFrom + flusher: // 26
		return (*rw26)(rw)

	case pusher + readerFrom + flusher + closeNotifier: // 27
		return (*rw27)(rw)

	case pusher + readerFrom + hijacker: // 28
		return (*rw28)(rw)

	case pusher + readerFrom + hijacker + closeNotifier: // 29
		return (*rw29)(rw)

	case pusher + readerFrom + hijacker + flusher: // 30
		return (*rw30)(rw)

	case pusher + readerFrom + hijacker + flusher + closeNotifier: // 31
		return (*rw31)(rw)

	default:
		panic(fmt.Errorf("unknown ResponseWriter index '%d'", index))
	}
}

/// ----------------------------------------------------------------------- ///

const (
	closeNotifier = 1 << iota
	flusher
	hijacker
	readerFrom
	pusher
)

func rwCloseNotify(rw *responseWriter) <-chan bool {
	return rw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

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
	_ ResponseWriter     = &rw1{}
	_ http.CloseNotifier = &rw1{} // 1
)

type rw1 responseWriter

func (w *rw1) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw1) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw1) Written() int64                             { return w.rw().Written() }
func (w *rw1) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw1) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw1) Header() http.Header                        { return w.rw().Header() }
func (w *rw1) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw1) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw1) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw1) CloseNotify() <-chan bool { return rwCloseNotify(w.rw()) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw2{}
	_ http.Flusher   = &rw2{} // 2
)

type rw2 responseWriter

func (w *rw2) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw2) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw2) Written() int64                             { return w.rw().Written() }
func (w *rw2) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw2) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw2) Header() http.Header                        { return w.rw().Header() }
func (w *rw2) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw2) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw2) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw2) Flush() { rwFlush(w.rw()) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw3{}
	_ http.CloseNotifier = &rw3{} // 1
	_ http.Flusher       = &rw3{} // 2
)

type rw3 responseWriter

func (w *rw3) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw3) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw3) Written() int64                             { return w.rw().Written() }
func (w *rw3) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw3) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw3) Header() http.Header                        { return w.rw().Header() }
func (w *rw3) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw3) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw3) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw3) CloseNotify() <-chan bool { return rwCloseNotify(w.rw()) }
func (w *rw3) Flush()                   { rwFlush(w.rw()) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw4{}
	_ http.Hijacker  = &rw4{} // 4
)

type rw4 responseWriter

func (w *rw4) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw4) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw4) Written() int64                             { return w.rw().Written() }
func (w *rw4) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw4) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw4) Header() http.Header                        { return w.rw().Header() }
func (w *rw4) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw4) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw4) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw4) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw5{}
	_ http.CloseNotifier = &rw5{} // 1
	_ http.Hijacker      = &rw5{} // 4
)

type rw5 responseWriter

func (w *rw5) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw5) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw5) Written() int64                             { return w.rw().Written() }
func (w *rw5) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw5) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw5) Header() http.Header                        { return w.rw().Header() }
func (w *rw5) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw5) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw5) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw5) CloseNotify() <-chan bool                     { return rwCloseNotify(w.rw()) }
func (w *rw5) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw6{}
	_ http.Flusher   = &rw6{} // 2
	_ http.Hijacker  = &rw6{} // 4
)

type rw6 responseWriter

func (w *rw6) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw6) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw6) Written() int64                             { return w.rw().Written() }
func (w *rw6) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw6) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw6) Header() http.Header                        { return w.rw().Header() }
func (w *rw6) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw6) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw6) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw6) Flush()                                       { rwFlush(w.rw()) }
func (w *rw6) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw7{}
	_ http.CloseNotifier = &rw7{} // 1
	_ http.Flusher       = &rw7{} // 2
	_ http.Hijacker      = &rw7{} // 4
)

type rw7 responseWriter

func (w *rw7) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw7) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw7) Written() int64                             { return w.rw().Written() }
func (w *rw7) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw7) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw7) Header() http.Header                        { return w.rw().Header() }
func (w *rw7) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw7) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw7) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw7) CloseNotify() <-chan bool                     { return rwCloseNotify(w.rw()) }
func (w *rw7) Flush()                                       { rwFlush(w.rw()) }
func (w *rw7) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw8{}
	_ io.ReaderFrom  = &rw8{} // 8
)

type rw8 responseWriter

func (w *rw8) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw8) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw8) Written() int64                             { return w.rw().Written() }
func (w *rw8) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw8) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw8) Header() http.Header                        { return w.rw().Header() }
func (w *rw8) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw8) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw8) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw8) ReadFrom(r io.Reader) (int64, error) { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw9{}
	_ http.CloseNotifier = &rw9{} // 1
	_ io.ReaderFrom      = &rw9{} // 8
)

type rw9 responseWriter

func (w *rw9) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw9) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw9) Written() int64                             { return w.rw().Written() }
func (w *rw9) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw9) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw9) Header() http.Header                        { return w.rw().Header() }
func (w *rw9) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw9) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw9) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw9) CloseNotify() <-chan bool            { return rwCloseNotify(w.rw()) }
func (w *rw9) ReadFrom(r io.Reader) (int64, error) { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw10{}
	_ http.Flusher   = &rw10{} // 2
	_ io.ReaderFrom  = &rw10{} // 8
)

type rw10 responseWriter

func (w *rw10) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw10) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw10) Written() int64                             { return w.rw().Written() }
func (w *rw10) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw10) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw10) Header() http.Header                        { return w.rw().Header() }
func (w *rw10) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw10) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw10) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw10) Flush()                              { rwFlush(w.rw()) }
func (w *rw10) ReadFrom(r io.Reader) (int64, error) { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw11{}
	_ http.CloseNotifier = &rw11{} // 1
	_ http.Flusher       = &rw11{} // 2
	_ io.ReaderFrom      = &rw11{} // 8
)

type rw11 responseWriter

func (w *rw11) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw11) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw11) Written() int64                             { return w.rw().Written() }
func (w *rw11) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw11) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw11) Header() http.Header                        { return w.rw().Header() }
func (w *rw11) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw11) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw11) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw11) CloseNotify() <-chan bool            { return rwCloseNotify(w.rw()) }
func (w *rw11) Flush()                              { rwFlush(w.rw()) }
func (w *rw11) ReadFrom(r io.Reader) (int64, error) { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw12{}
	_ http.Hijacker  = &rw12{} // 4
	_ io.ReaderFrom  = &rw12{} // 8
)

type rw12 responseWriter

func (w *rw12) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw12) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw12) Written() int64                             { return w.rw().Written() }
func (w *rw12) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw12) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw12) Header() http.Header                        { return w.rw().Header() }
func (w *rw12) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw12) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw12) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw12) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }
func (w *rw12) ReadFrom(r io.Reader) (int64, error)          { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw13{}
	_ http.CloseNotifier = &rw13{} // 1
	_ http.Hijacker      = &rw13{} // 4
	_ io.ReaderFrom      = &rw13{} // 8
)

type rw13 responseWriter

func (w *rw13) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw13) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw13) Written() int64                             { return w.rw().Written() }
func (w *rw13) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw13) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw13) Header() http.Header                        { return w.rw().Header() }
func (w *rw13) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw13) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw13) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw13) CloseNotify() <-chan bool                     { return rwCloseNotify(w.rw()) }
func (w *rw13) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }
func (w *rw13) ReadFrom(r io.Reader) (int64, error)          { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw14{}
	_ http.Flusher   = &rw14{} // 2
	_ http.Hijacker  = &rw14{} // 4
	_ io.ReaderFrom  = &rw14{} // 8
)

type rw14 responseWriter

func (w *rw14) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw14) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw14) Written() int64                             { return w.rw().Written() }
func (w *rw14) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw14) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw14) Header() http.Header                        { return w.rw().Header() }
func (w *rw14) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw14) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw14) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw14) Flush()                                       { rwFlush(w.rw()) }
func (w *rw14) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }
func (w *rw14) ReadFrom(r io.Reader) (int64, error)          { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw15{}
	_ http.CloseNotifier = &rw15{} // 1
	_ http.Flusher       = &rw15{} // 2
	_ http.Hijacker      = &rw15{} // 4
	_ io.ReaderFrom      = &rw15{} // 8
)

type rw15 responseWriter

func (w *rw15) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw15) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw15) Written() int64                             { return w.rw().Written() }
func (w *rw15) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw15) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw15) Header() http.Header                        { return w.rw().Header() }
func (w *rw15) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw15) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw15) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw15) CloseNotify() <-chan bool                     { return rwCloseNotify(w.rw()) }
func (w *rw15) Flush()                                       { rwFlush(w.rw()) }
func (w *rw15) Hijack() (net.Conn, *bufio.ReadWriter, error) { return rwHijack(w.rw()) }
func (w *rw15) ReadFrom(r io.Reader) (int64, error)          { return rwReadFrom(w.rw(), r) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw16{}
	_ http.Pusher    = &rw16{} // 16
)

type rw16 responseWriter

func (w *rw16) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw16) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw16) Written() int64                             { return w.rw().Written() }
func (w *rw16) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw16) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw16) Header() http.Header                        { return w.rw().Header() }
func (w *rw16) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw16) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw16) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw16) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw17{}
	_ http.CloseNotifier = &rw17{} // 1
	_ http.Pusher        = &rw17{} // 16
)

type rw17 responseWriter

func (w *rw17) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw17) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw17) Written() int64                             { return w.rw().Written() }
func (w *rw17) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw17) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw17) Header() http.Header                        { return w.rw().Header() }
func (w *rw17) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw17) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw17) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw17) CloseNotify() <-chan bool                         { return rwCloseNotify(w.rw()) }
func (w *rw17) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw18{}
	_ http.Flusher   = &rw18{} // 2
	_ http.Pusher    = &rw18{} // 16
)

type rw18 responseWriter

func (w *rw18) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw18) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw18) Written() int64                             { return w.rw().Written() }
func (w *rw18) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw18) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw18) Header() http.Header                        { return w.rw().Header() }
func (w *rw18) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw18) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw18) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw18) Flush()                                           { rwFlush(w.rw()) }
func (w *rw18) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw19{}
	_ http.CloseNotifier = &rw19{} // 1
	_ http.Flusher       = &rw19{} // 2
	_ http.Pusher        = &rw19{} // 16
)

type rw19 responseWriter

func (w *rw19) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw19) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw19) Written() int64                             { return w.rw().Written() }
func (w *rw19) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw19) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw19) Header() http.Header                        { return w.rw().Header() }
func (w *rw19) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw19) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw19) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw19) CloseNotify() <-chan bool                         { return rwCloseNotify(w.rw()) }
func (w *rw19) Flush()                                           { rwFlush(w.rw()) }
func (w *rw19) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw20{}
	_ http.Hijacker  = &rw20{} // 4
	_ http.Pusher    = &rw20{} // 16
)

type rw20 responseWriter

func (w *rw20) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw20) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw20) Written() int64                             { return w.rw().Written() }
func (w *rw20) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw20) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw20) Header() http.Header                        { return w.rw().Header() }
func (w *rw20) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw20) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw20) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw20) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw20) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw21{}
	_ http.CloseNotifier = &rw21{} // 1
	_ http.Hijacker      = &rw21{} // 4
	_ http.Pusher        = &rw21{} // 16
)

type rw21 responseWriter

func (w *rw21) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw21) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw21) Written() int64                             { return w.rw().Written() }
func (w *rw21) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw21) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw21) Header() http.Header                        { return w.rw().Header() }
func (w *rw21) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw21) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw21) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw21) CloseNotify() <-chan bool                         { return rwCloseNotify(w.rw()) }
func (w *rw21) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw21) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw22{}
	_ http.Flusher   = &rw22{} // 2
	_ http.Hijacker  = &rw22{} // 4
	_ http.Pusher    = &rw22{} // 16
)

type rw22 responseWriter

func (w *rw22) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw22) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw22) Written() int64                             { return w.rw().Written() }
func (w *rw22) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw22) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw22) Header() http.Header                        { return w.rw().Header() }
func (w *rw22) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw22) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw22) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw22) Flush()                                           { rwFlush(w.rw()) }
func (w *rw22) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw22) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw23{}
	_ http.CloseNotifier = &rw23{} // 1
	_ http.Flusher       = &rw23{} // 2
	_ http.Hijacker      = &rw23{} // 4
	_ http.Pusher        = &rw23{} // 16
)

type rw23 responseWriter

func (w *rw23) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw23) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw23) Written() int64                             { return w.rw().Written() }
func (w *rw23) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw23) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw23) Header() http.Header                        { return w.rw().Header() }
func (w *rw23) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw23) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw23) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw23) CloseNotify() <-chan bool                         { return rwCloseNotify(w.rw()) }
func (w *rw23) Flush()                                           { rwFlush(w.rw()) }
func (w *rw23) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw23) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw24{}
	_ io.ReaderFrom  = &rw24{} // 8
	_ http.Pusher    = &rw24{} // 16
)

type rw24 responseWriter

func (w *rw24) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw24) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw24) Written() int64                             { return w.rw().Written() }
func (w *rw24) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw24) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw24) Header() http.Header                        { return w.rw().Header() }
func (w *rw24) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw24) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw24) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw24) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw24) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw25{}
	_ http.CloseNotifier = &rw25{} // 1
	_ io.ReaderFrom      = &rw25{} // 8
	_ http.Pusher        = &rw25{} // 16
)

type rw25 responseWriter

func (w *rw25) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw25) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw25) Written() int64                             { return w.rw().Written() }
func (w *rw25) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw25) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw25) Header() http.Header                        { return w.rw().Header() }
func (w *rw25) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw25) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw25) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw25) CloseNotify() <-chan bool                         { return rwCloseNotify(w.rw()) }
func (w *rw25) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw25) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw26{}
	_ http.Flusher   = &rw26{} // 2
	_ io.ReaderFrom  = &rw26{} // 8
	_ http.Pusher    = &rw26{} // 16
)

type rw26 responseWriter

func (w *rw26) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw26) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw26) Written() int64                             { return w.rw().Written() }
func (w *rw26) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw26) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw26) Header() http.Header                        { return w.rw().Header() }
func (w *rw26) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw26) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw26) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw26) Flush()                                           { rwFlush(w.rw()) }
func (w *rw26) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw26) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw27{}
	_ http.CloseNotifier = &rw27{} // 1
	_ http.Flusher       = &rw27{} // 2
	_ io.ReaderFrom      = &rw27{} // 8
	_ http.Pusher        = &rw27{} // 16
)

type rw27 responseWriter

func (w *rw27) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw27) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw27) Written() int64                             { return w.rw().Written() }
func (w *rw27) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw27) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw27) Header() http.Header                        { return w.rw().Header() }
func (w *rw27) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw27) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw27) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw27) CloseNotify() <-chan bool                         { return rwCloseNotify(w.rw()) }
func (w *rw27) Flush()                                           { rwFlush(w.rw()) }
func (w *rw27) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw27) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw28{}
	_ http.Hijacker  = &rw28{} // 4
	_ io.ReaderFrom  = &rw28{} // 8
	_ http.Pusher    = &rw28{} // 16
)

type rw28 responseWriter

func (w *rw28) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw28) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw28) Written() int64                             { return w.rw().Written() }
func (w *rw28) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw28) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw28) Header() http.Header                        { return w.rw().Header() }
func (w *rw28) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw28) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw28) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw28) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw28) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw28) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw29{}
	_ http.CloseNotifier = &rw29{} // 1
	_ http.Hijacker      = &rw29{} // 4
	_ io.ReaderFrom      = &rw29{} // 8
	_ http.Pusher        = &rw29{} // 16
)

type rw29 responseWriter

func (w *rw29) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw29) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw29) Written() int64                             { return w.rw().Written() }
func (w *rw29) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw29) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw29) Header() http.Header                        { return w.rw().Header() }
func (w *rw29) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw29) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw29) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw29) CloseNotify() <-chan bool                         { return rwCloseNotify(w.rw()) }
func (w *rw29) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw29) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw29) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter = &rw30{}
	_ http.Flusher   = &rw30{} // 2
	_ http.Hijacker  = &rw30{} // 4
	_ io.ReaderFrom  = &rw30{} // 8
	_ http.Pusher    = &rw30{} // 16
)

type rw30 responseWriter

func (w *rw30) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw30) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw30) Written() int64                             { return w.rw().Written() }
func (w *rw30) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw30) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw30) Header() http.Header                        { return w.rw().Header() }
func (w *rw30) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw30) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw30) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw30) Flush()                                           { rwFlush(w.rw()) }
func (w *rw30) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw30) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw30) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }

/// ----------------------------------------------------------------------- ///

var (
	_ ResponseWriter     = &rw31{}
	_ http.CloseNotifier = &rw31{} // 1
	_ http.Flusher       = &rw31{} // 2
	_ http.Hijacker      = &rw31{} // 4
	_ io.ReaderFrom      = &rw31{} // 8
	_ http.Pusher        = &rw31{} // 16
)

type rw31 responseWriter

func (w *rw31) rw() *responseWriter                        { return (*responseWriter)(w) }
func (w *rw31) WrappedResponseWriter() http.ResponseWriter { return w.rw().ResponseWriter }
func (w *rw31) Written() int64                             { return w.rw().Written() }
func (w *rw31) StatusCode() int                            { return w.rw().StatusCode() }
func (w *rw31) WroteHeader() bool                          { return w.rw().WroteHeader() }
func (w *rw31) Header() http.Header                        { return w.rw().Header() }
func (w *rw31) Write(p []byte) (int, error)                { return w.rw().Write(p) }
func (w *rw31) WriteString(s string) (int, error)          { return w.rw().WriteString(s) }
func (w *rw31) WriteHeader(code int)                       { w.rw().WriteHeader(code) }

func (w *rw31) CloseNotify() <-chan bool                         { return rwCloseNotify(w.rw()) }
func (w *rw31) Flush()                                           { rwFlush(w.rw()) }
func (w *rw31) Hijack() (net.Conn, *bufio.ReadWriter, error)     { return rwHijack(w.rw()) }
func (w *rw31) ReadFrom(r io.Reader) (int64, error)              { return rwReadFrom(w.rw(), r) }
func (w *rw31) Push(target string, opts *http.PushOptions) error { return rwPush(w.rw(), target, opts) }
