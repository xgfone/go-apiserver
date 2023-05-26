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
	"fmt"
	"net"
	"net/http"

	"github.com/xgfone/go-apiserver/http/router"
	"github.com/xgfone/go-apiserver/nets/stream"
)

func init() {
	registerHTTPServerBuilder("http", "tcp")
	registerHTTPServerBuilder("http4", "tcp4")
	registerHTTPServerBuilder("http6", "tcp6")
	registerHTTPServerBuilder("unixhttp", "unix")
}

func registerHTTPServerBuilder(proto, streamProto string) {
	RegisterServerBuilder(proto, func(addr string, h interface{}) (Server, error) {
		ln, err := stream.Listen(streamProto, addr)
		if err != nil {
			return nil, err
		}

		var httpHandler http.Handler
		switch handler := h.(type) {
		case nil:
		case http.Handler:
			httpHandler = handler
		default:
			panic(fmt.Errorf("unknown http handler type '%T'", h))
		}

		server := NewHTTPServer(ln, httpHandler)
		server.Proto = proto
		return server, nil
	})
}

var _ Server = HTTPServer{}

// HTTPServer represents a http entrypoint server.
type HTTPServer struct {
	StreamServer
	HTTPHandler http.Handler
	HTTPServer  *stream.HTTPServerHandler
}

// NewHTTPServer returns a new HTTP entrypoint Server based on TCP.
func NewHTTPServer(ln net.Listener, handler http.Handler) (server HTTPServer) {
	if handler == nil {
		handler = router.DefaultRouter
	}

	server.HTTPHandler = handler
	server.HTTPServer = stream.NewHTTPServerHandler(ln.Addr(), handler)
	server.StreamServer = NewStreamServer("http", ln, server.HTTPServer)
	return
}

// Start starts and runs the http server until it is closed.
func (s HTTPServer) Start() {
	go s.HTTPServer.Start()
	s.StreamServer.Start()
}
