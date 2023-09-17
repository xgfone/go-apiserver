// Copyright 2022~2023 xgfone
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

	"github.com/xgfone/go-apiserver/middleware"
	"github.com/xgfone/go-apiserver/nets/stream"
)

func init() {
	// TCP
	RegisterServerBuilder("tcp", func(addr string, h interface{}) (Server, error) {
		ln, err := stream.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}
		return NewStreamServer("tcp", ln, h.(stream.Handler)), nil
	})

	// Unix
	RegisterServerBuilder("unix", func(addr string, h interface{}) (Server, error) {
		ln, err := stream.Listen("unix", addr)
		if err != nil {
			return nil, err
		}
		return NewStreamServer("unix", ln, h.(stream.Handler)), nil
	})
}

var _ Server = StreamServer{}

// StreamServer represents a stream entrypoint server.
type StreamServer struct {
	Proto       string
	Middlewares *middleware.Manager
	*stream.Server
}

// NewStreamServer returns a new stream entrypoint Server, such as TCP.
func NewStreamServer(proto string, ln net.Listener, handler stream.Handler) (server StreamServer) {
	if ln == nil {
		panic("the stream listener is nil")
	} else if handler == nil {
		panic("the stream handler is nil")
	}

	server.Proto = proto
	server.Middlewares = middleware.NewManager(handler)
	server.Server = stream.NewServer(ln, server.Middlewares)
	return
}

// Addr returns the address that the server listens on.
func (s StreamServer) Addr() net.Addr { return s.Server.Listener.Addr() }

// Protocol returns the protocol of the stream server.
func (s StreamServer) Protocol() string { return s.Proto }
