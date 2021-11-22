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

package tcp

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
)

type testHandler struct {
	buf     *bytes.Buffer
	name    string
	handler Handler
}

func (h testHandler) OnShutdown(c context.Context) {
	if h.handler == nil {
		fmt.Fprintf(h.buf, "'%s' onshutdown handler\n", h.name)
		return
	}

	fmt.Fprintf(h.buf, "'%s' onshutdown before middleware\n", h.name)
	h.handler.OnShutdown(c)
	fmt.Fprintf(h.buf, "'%s' onshutdown after middleware\n", h.name)
}

func (h testHandler) OnServerExit(err error) {
	if h.handler == nil {
		fmt.Fprintf(h.buf, "'%s' onserverexit handler\n", h.name)
		return
	}

	fmt.Fprintf(h.buf, "'%s' onserverexit before middleware\n", h.name)
	h.handler.OnServerExit(err)
	fmt.Fprintf(h.buf, "'%s' onserverexit after middleware\n", h.name)
}

func (h testHandler) OnConnection(c net.Conn) {
	if h.handler == nil {
		fmt.Fprintf(h.buf, "'%s' onconnection handler\n", h.name)
		return
	}

	fmt.Fprintf(h.buf, "'%s' onconnection before middleware\n", h.name)
	h.handler.OnConnection(c)
	fmt.Fprintf(h.buf, "'%s' onconnection after middleware\n", h.name)
}

func handlerMiddleware(name string, buf *bytes.Buffer) Middleware {
	return func(h Handler) Handler {
		return testHandler{buf: buf, name: name, handler: h}
	}
}

func TestMiddleware(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	handler := NewMiddlewareHandler(testHandler{buf: buf, name: "h"},
		handlerMiddleware("mw1", buf),
		handlerMiddleware("mw2", buf),
		handlerMiddleware("mw3", buf),
	)

	handler.OnConnection(nil)
	handler.OnServerExit(nil)
	handler.OnShutdown(context.Background())

	expectedResults := []string{
		// OnConnection
		"'mw1' onconnection before middleware",
		"'mw2' onconnection before middleware",
		"'mw3' onconnection before middleware",
		"'h' onconnection handler",
		"'mw3' onconnection after middleware",
		"'mw2' onconnection after middleware",
		"'mw1' onconnection after middleware",

		// OnServerExit
		"'mw1' onserverexit before middleware",
		"'mw2' onserverexit before middleware",
		"'mw3' onserverexit before middleware",
		"'h' onserverexit handler",
		"'mw3' onserverexit after middleware",
		"'mw2' onserverexit after middleware",
		"'mw1' onserverexit after middleware",

		// OnShutdown
		"'mw1' onshutdown before middleware",
		"'mw2' onshutdown before middleware",
		"'mw3' onshutdown before middleware",
		"'h' onshutdown handler",
		"'mw3' onshutdown after middleware",
		"'mw2' onshutdown after middleware",
		"'mw1' onshutdown after middleware",

		// End
		"",
	}
	lines := strings.Split(buf.String(), "\n")
	if len(lines) != len(expectedResults) {
		t.Errorf("expect %d lines, but got %d lines", len(expectedResults), len(lines))
	} else {
		for i, line := range expectedResults {
			if lines[i] != line {
				t.Errorf("%d line: expect '%s', but got '%s'", i, line, lines[i])
			}
		}
	}
}
