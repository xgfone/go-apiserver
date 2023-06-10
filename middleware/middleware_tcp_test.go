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

package middleware

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/internal/test"
	"github.com/xgfone/go-apiserver/nets/stream"
)

var _ stream.Handler = testTCPHandler{}

type testTCPHandler struct {
	buf     *bytes.Buffer
	name    string
	handler stream.Handler
}

func (h testTCPHandler) OnServerExit(err error) {
	if h.handler == nil {
		fmt.Fprintf(h.buf, "'%s' onserverexit handler\n", h.name)
		return
	}

	fmt.Fprintf(h.buf, "'%s' onserverexit before middleware\n", h.name)
	h.handler.OnServerExit(err)
	fmt.Fprintf(h.buf, "'%s' onserverexit after middleware\n", h.name)
}

func (h testTCPHandler) OnConnection(c net.Conn) {
	if h.handler == nil {
		fmt.Fprintf(h.buf, "'%s' onconnection handler\n", h.name)
		return
	}

	fmt.Fprintf(h.buf, "'%s' onconnection before middleware\n", h.name)
	h.handler.OnConnection(c)
	fmt.Fprintf(h.buf, "'%s' onconnection after middleware\n", h.name)
}

func tcpMiddleware(name string, priority int, buf *bytes.Buffer) Middleware {
	return NewMiddleware(name, priority, func(h interface{}) interface{} {
		return testTCPHandler{handler: h.(stream.Handler), name: name, buf: buf}
	})
}

func TestMiddlewareManagerTCP(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	tcpHandler := testTCPHandler{buf: buf, name: "handler"}

	manager := NewManager(nil)
	manager.Use(tcpMiddleware("mw2", 2, buf), tcpMiddleware("mw1", 1, buf))

	handler := manager.WrapHandler(tcpHandler).(stream.Handler)
	handler.OnConnection(nil)
	handler.OnServerExit(nil)

	expects := []string{
		// OnConnection
		"'mw1' onconnection before middleware",
		"'mw2' onconnection before middleware",
		"'handler' onconnection handler",
		"'mw2' onconnection after middleware",
		"'mw1' onconnection after middleware",

		// OnServerExit
		"'mw1' onserverexit before middleware",
		"'mw2' onserverexit before middleware",
		"'handler' onserverexit handler",
		"'mw2' onserverexit after middleware",
		"'mw1' onserverexit after middleware",

		// End
		"",
	}
	test.CheckStrings(t, "MiddlewareManagerTCP", strings.Split(buf.String(), "\n"), expects)

	buf.Reset()
	manager.SetHandler(tcpHandler)
	manager.OnConnection(nil)
	manager.OnServerExit(nil)
	test.CheckStrings(t, "MiddlewareManagerTCP", strings.Split(buf.String(), "\n"), expects)
}
