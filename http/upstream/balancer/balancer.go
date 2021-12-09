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
	"net/http"

	"github.com/xgfone/go-apiserver/http/upstream"
)

// Balancer is used to forward the request to one of the backend servers.
type Balancer interface {
	Forward(http.ResponseWriter, *http.Request, upstream.Servers) error
	Policy() string
}

/// ---------------------------------------------------------------------- ///

var balancers = make(map[string]Balancer, 16)

func init() {
	Register(Random())
	Register(RoundRobin())
	Register(WeightedRandom())
	Register(WeightedRoundRobin())
	Register(SourceIPHash())
	Register(LeastConn())
}

// Register registers the given balancer.
//
// If the registering balancer policy has existed, override it to the new.
func Register(balancer Balancer) { balancers[balancer.Policy()] = balancer }

// Get returns the registered balancer by the balancer policy.
//
// If the balancer policy does not exist, return nil.
func Get(policy string) Balancer { return balancers[policy] }

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
