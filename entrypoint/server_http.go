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

package entrypoint

import (
	"net"
	"net/http"

	"github.com/xgfone/go-apiserver/http/router"
	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
	"github.com/xgfone/go-apiserver/tcp"
)

var _ Server = HTTPServer{}

// HTTPServer represents a http entrypoint server.
type HTTPServer struct {
	HTTPHandler http.Handler
	HTTPServer  *tcp.HTTPServerHandler
	TCPServer
}

// NewHTTPServer returns a new HTTP entrypoint Server.
func NewHTTPServer(serverName string, ln net.Listener, handler http.Handler) (server HTTPServer) {
	if handler == nil {
		handler = router.NewRouter(ruler.NewRouteManager())
	}

	server.HTTPHandler = handler
	server.HTTPServer = tcp.NewHTTPServerHandler(ln.Addr(), handler)
	server.TCPServer = NewTCPServer(serverName, ln, server.HTTPServer)
	return
}

// Protocal returns the protocal of the http server, which is a fixed "http".
func (s HTTPServer) Protocal() string { return "http" }

// Start starts and runs the http server until it is closed.
func (s HTTPServer) Start() {
	go s.HTTPServer.Start()
	s.TCPServer.Start()
}
