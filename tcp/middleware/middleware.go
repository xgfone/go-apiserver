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

// Package middleware provides the middleware functions for the tcp handler.
package middleware

import (
	"errors"

	"github.com/xgfone/go-apiserver/nets"
	"github.com/xgfone/go-apiserver/tcp"
)

// Middleware is the tcp handler middleware.
type Middleware interface {
	// Name returns the name of the middleware.
	Name() string

	// The smaller the value, the higher the priority and the middleware
	// is executed preferentially.
	Priority() int

	// TCPHandler is used to wrap the tcp handler and returns a new one.
	TCPHandler(tcp.Handler) tcp.Handler
}

type middleware struct {
	name     string
	priority int
	handler  func(tcp.Handler) tcp.Handler
}

func (m middleware) Name() string                         { return m.name }
func (m middleware) Priority() int                        { return m.priority }
func (m middleware) TCPHandler(h tcp.Handler) tcp.Handler { return m.handler(h) }

// NewMiddleware returns a new TCP handler middleware.
func NewMiddleware(name string, prio int, f func(tcp.Handler) tcp.Handler) Middleware {
	return middleware{name: name, priority: prio, handler: f}
}

// Middlewares is a group of the tcp handler middlewares.
type Middlewares []Middleware

func (ms Middlewares) Len() int           { return len(ms) }
func (ms Middlewares) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms Middlewares) Less(i, j int) bool { return ms[i].Priority() < ms[j].Priority() }

// Handler wraps the tcp handler with the middlewares and returns a new one.
func (ms Middlewares) Handler(handler tcp.Handler) tcp.Handler {
	for _len := len(ms) - 1; _len >= 0; _len-- {
		handler = ms[_len].TCPHandler(handler)
	}
	return handler
}

// Clone clones itself and appends the new middlewares to the new.
func (ms Middlewares) Clone(news ...Middleware) Middlewares {
	mws := make(Middlewares, len(ms)+len(news))
	copy(mws, ms)
	copy(mws[len(ms):], news)
	return mws
}

// IPWhitelist returns a tcp middleware to filter the connections
// that the client ip is not in the given ip or cidr list.
func IPWhitelist(priority int, ipOrCidrs ...string) (Middleware, error) {
	if len(ipOrCidrs) == 0 {
		return nil, errors.New("TCP ClientIP middleware: no ips or cidrs")
	}

	checker, err := nets.NewIPCheckers(ipOrCidrs...)
	if err != nil {
		return nil, err
	}

	return NewMiddleware("ip_whitelist", priority, func(h tcp.Handler) tcp.Handler {
		return tcp.NewIPWhitelistHandler(h, checker)
	}), nil
}
