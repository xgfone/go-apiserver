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

package balancer

import (
	"math"
	"net/http"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/http/upstream"
)

func init() {
	registerBuiltinBuidler("round_robin", RoundRobin)
	registerBuiltinBuidler("weight_round_robin", WeightedRoundRobin)
}

// RoundRobin returns a new balancer based on the roundrobin.
//
// The policy name is "round_robin".
func RoundRobin() Balancer {
	last := uint64(math.MaxUint64)
	return NewBalancer("round_robin",
		func(w http.ResponseWriter, r *http.Request, s upstream.Servers) error {
			pos := atomic.AddUint64(&last, 1)
			return forward(w, r, s[pos%uint64(len(s))])
		})
}

// WeightedRoundRobin returns a new balancer based on the roundrobin and weight.
//
// The policy name is "weight_round_robin".
func WeightedRoundRobin() Balancer {
	last := uint64(math.MaxUint64)
	return NewBalancer("weight_round_robin",
		func(w http.ResponseWriter, r *http.Request, s upstream.Servers) error {
			pos := atomic.AddUint64(&last, 1)
			return forward(w, r, calcServerOnWeight(s, pos))
		})
}
