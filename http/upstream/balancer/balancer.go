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

// Package balancer implements some balancer forwarding policies.
package balancer

import (
	"fmt"
	"net/http"

	"github.com/xgfone/go-apiserver/http/upstream"
)

// SelectedServerCallback is a callback for the server selected by the balancer
// to forward the request to.
type SelectedServerCallback func(req *http.Request, selectedServer upstream.Server)

// Balancer is used to forward the request to one of the backend servers.
type Balancer interface {
	Forward(http.ResponseWriter, *http.Request, upstream.Servers) error
	Policy() string
}

func serverCallback(callback SelectedServerCallback, w http.ResponseWriter,
	r *http.Request, s upstream.Server) (err error) {
	if err = s.HandleHTTP(w, r); err == nil && callback != nil {
		callback(r, s)
	}
	return
}

/// ---------------------------------------------------------------------- ///

// Builder is used to build a new Balancer with the config.
type Builder func(config interface{}) (Balancer, error)

var builders = make(map[string]Builder, 16)

func registerBuiltinBuidler(t string, f func(SelectedServerCallback) Balancer) {
	const s = `balancer builder typed '%s' needs the type SelectedServerCallback, but got '%T'`
	RegisterBuidler(t, func(config interface{}) (balancer Balancer, err error) {
		switch v := config.(type) {
		case nil:
			balancer = f(nil)

		case SelectedServerCallback:
			balancer = f(v)

		case func(*http.Request, upstream.Server):
			balancer = f(v)

		default:
			err = fmt.Errorf(s, t, config)
		}

		return
	})
}

// RegisterBuidler registers the given balancer builder.
//
// If the balancer builder typed "typ" has existed, override it to the new.
func RegisterBuidler(typ string, builder Builder) { builders[typ] = builder }

// GetBuilder returns the registered balancer builder by the type.
//
// If the balancer builder typed "typ" does not exist, return nil.
func GetBuilder(typ string) Builder { return builders[typ] }

// Build is a convenient function to build a new balancer typed "typ".
func Build(typ string, config interface{}) (balancer Balancer, err error) {
	if builder := GetBuilder(typ); builder != nil {
		balancer, err = builder(config)
	} else {
		err = fmt.Errorf("no the builder typed '%s'", typ)
	}
	return
}

/// ---------------------------------------------------------------------- ///

// ForwardFunc is the function to forward the request to one of the servers.
type ForwardFunc func(http.ResponseWriter, *http.Request, upstream.Servers) error

type servers = upstream.Servers

type forwarder struct {
	forward ForwardFunc
	policy  string
}

func (f forwarder) Policy() string { return f.policy }
func (f forwarder) Forward(w http.ResponseWriter, r *http.Request, s servers) error {
	return f.forward(w, r, s)
}

// NewForwarder returns a new forwarder with the policy and the forwarder function.
func NewForwarder(policy string, forward ForwardFunc) Balancer {
	return forwarder{forward: forward, policy: policy}
}

/// ---------------------------------------------------------------------- ///

type retry struct{ Balancer }

func (f retry) WrappedBalancer() Balancer { return f.Balancer }
func (f retry) Forward(w http.ResponseWriter, r *http.Request, s servers) (err error) {
	_len := len(s)
	if _len == 1 {
		return s[0].HandleHTTP(w, r)
	}

	for ; _len > 0; _len-- {
		if err = f.Balancer.Forward(w, r, s); err == nil {
			break
		}
	}

	return
}

// Retry returns a new balancer to retry the rest servers when failing to
// forward the request.
func Retry(balancer Balancer) Balancer {
	if balancer == nil {
		panic("RetryBalancer: the wrapped balancer is nil")
	}
	return retry{Balancer: balancer}
}

/// ---------------------------------------------------------------------- ///

func calcServerOnWeight(ss servers, pos uint64) upstream.Server {
	_len := len(ss) - 1
	for {
		var total uint64
		for i := _len; i >= 0; i-- {
			total += uint64(upstream.GetServerWeight(ss[i]))
			if pos < total {
				return ss[i]
			}
		}
		pos %= total
	}
}
