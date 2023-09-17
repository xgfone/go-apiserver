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

// Package server provides a common http server.
package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/xgfone/go-atexit"
)

// New returns a new http server with the handler.
func New(addr string, handler http.Handler) *http.Server {
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

// Start just starts the http server, which is equal to
//
//	StartWithListenerCallback(server, nil)
func Start(server *http.Server) { StartWithListenerCallback(server, nil) }

// StartWithListenerCallback starts the http server until it is stopped.
func StartWithListenerCallback(server *http.Server, cb func(net.Listener) net.Listener) {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		slog.Error("fail to open the listener on the address",
			"protocol", "tcp", "addr", server.Addr, "err", err)
		return
	}

	if cb != nil {
		ln = cb(ln)
	}

	slog.Info("start the http server", "addr", server.Addr)
	defer slog.Info("stop the http server", "addr", server.Addr)

	server.RegisterOnShutdown(atexit.Execute)
	atexit.OnExit(func() { Stop(server) })
	_ = server.Serve(ln)
	atexit.Wait()
}

// Stop just stops the http server, which is equal to
//
//	StopWithContext(context.Background(), server)
func Stop(server *http.Server) {
	StopWithContext(context.Background(), server)
}

// StopWithContext stops the http server.
func StopWithContext(ctx context.Context, server *http.Server) {
	_ = server.Shutdown(ctx)
}
