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

package loadbalancer

import (
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/xgfone/go-apiserver/http/upstream"
)

// Forwarder is used to forward the request to one of the backend servers.
type Forwarder interface {
	Forward(http.ResponseWriter, *http.Request, upstream.Servers) error
	Policy() string
}

// ForwardFunc is the function to forward the request to one of the servers.
type ForwardFunc func(http.ResponseWriter, *http.Request, upstream.Servers) error

type forwarder struct {
	forward ForwardFunc
	policy  string
}

func (f forwarder) Policy() string { return f.policy }
func (f forwarder) Forward(w http.ResponseWriter, r *http.Request, s upstream.Servers) error {
	return f.forward(w, r, s)
}

// NewForwarder returns a new forwarder with the policy and the forwarder function.
func NewForwarder(policy string, forward ForwardFunc) Forwarder {
	return forwarder{forward: forward, policy: policy}
}

type retry struct{ Forwarder }

func (f retry) WrappedForwarder() Forwarder { return f.Forwarder }
func (f retry) Forward(w http.ResponseWriter, r *http.Request, s upstream.Servers) (err error) {
	for _len := len(s); _len > 0; _len-- {
		if err = f.Forwarder.Forward(w, r, s); err == nil {
			break
		}
	}
	return
}

// Retry returns a new forwarder to retry the rest servers when failing to
// forward the request.
func Retry(forwarder Forwarder) Forwarder {
	if forwarder == nil {
		panic("RetryForwarder: the wrapped forwarder is nil")
	}
	return retry{Forwarder: forwarder}
}

// RoundRobin returns a new forwarder based on the roundrobin.
//
// The policy name is "round_robin".
func RoundRobin() Forwarder { return roundRobin(int(time.Now().UnixNano())) }

func roundRobin(start int) Forwarder {
	last := uint64(start)
	return NewForwarder("round_robin",
		func(w http.ResponseWriter, r *http.Request, s upstream.Servers) error {
			pos := atomic.AddUint64(&last, 1)
			return s[pos%uint64(len(s))].HandleHTTP(w, r)
		})
}

// Weight returns a new forwarder based on the weight.
//
// The policy name is "weight".
func Weight() Forwarder {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	getWeight := func(server upstream.Server) (weight int) {
		if ws, ok := server.(upstream.WeightedServer); ok {
			weight = ws.Weight()
		}
		return
	}

	return NewForwarder("weight",
		func(w http.ResponseWriter, r *http.Request, servers upstream.Servers) error {
			length := len(servers)
			sameWeight := true
			firstWeight := getWeight(servers[0])
			totalWeight := firstWeight

			weights := make([]int, length)
			weights[0] = firstWeight

			for i := 1; i < length; i++ {
				weight := getWeight(servers[i])
				weights[i] = weight
				totalWeight += weight
				if sameWeight && weight != firstWeight {
					sameWeight = false
				}
			}

			if !sameWeight && totalWeight > 0 {
				offset := random.Intn(totalWeight)
				for i := 0; i < length; i++ {
					if offset -= weights[i]; offset < 0 {
						return servers[i].HandleHTTP(w, r)
					}
				}
			}

			return servers[random.Intn(len(servers))].HandleHTTP(w, r)
		})
}
