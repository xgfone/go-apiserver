// Copyright 2023 xgfone
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

// Package server provides a simple common http server starter.
package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/xgfone/go-apiserver/http/router"
	"github.com/xgfone/go-defaults"
)

// ServeWithListener is used to start the http server with listener
// until it is stopped.
var ServeWithListener func(server *http.Server, ln net.Listener)

// New returns a new http server with the handler.
//
// If handler is nil, use router.DefaultRouter instead.
func New(addr string, handler http.Handler) *http.Server {
	if handler == nil {
		handler = router.DefaultRouter
	}

	return &http.Server{
		Addr:    addr,
		Handler: handler,

		ReadTimeout:  0,
		WriteTimeout: 0,

		IdleTimeout:       time.Minute * 3,
		ReadHeaderTimeout: time.Second * 3,

		ErrorLog: slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}
}

// Serve starts the http server with server.Addr until it is stopped.
func Serve(server *http.Server) {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		slog.Error("fail to open the listener on the address",
			"protocol", "tcp", "addr", server.Addr, "err", err)
		return
	}

	if ServeWithListener != nil {
		ServeWithListener(server, ln)
	} else {
		DefaultServeWithListener(server, ln)
	}
}

// Start is a convenient function to start the http server with addr and handler.
func Start(addr string, handler http.Handler) {
	Serve(New(addr, handler))
}

// Stop is a convenient function to stop the http server.
func Stop(server *http.Server) {
	_ = server.Shutdown(context.Background())
}

// DefaultServeWithListener is the default implementation to start the http server.
func DefaultServeWithListener(server *http.Server, ln net.Listener) {
	defaults.OnExit(func() { Stop(server) })
	serve(server, ln)
	defaults.Exit(0)
}

func serve(server *http.Server, ln net.Listener) {
	slog.Info("start the http server", "addr", server.Addr)
	defer slog.Info("stop the http server", "addr", server.Addr)
	_ = server.Serve(ln)
}
