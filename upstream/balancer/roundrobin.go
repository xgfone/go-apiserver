// Copyright 2021~2023 xgfone
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
	"context"
	"math"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/upstream"
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
		func(c context.Context, r interface{}, sd upstream.ServerDiscovery) error {
			ss := sd.OnServers()
			_len := len(ss)
			if _len == 1 {
				return ss[0].Serve(c, r)
			}

			pos := atomic.AddUint64(&last, 1)
			return ss[pos%uint64(_len)].Serve(c, r)
		})
}

// WeightedRoundRobin returns a new balancer based on the roundrobin and weight.
//
// The policy name is "weight_round_robin".
func WeightedRoundRobin() Balancer {
	ctx := &weightedRRServerCtx{caches: make(map[string]*weightedRRServer, 16)}
	return NewBalancer("weight_round_robin",
		func(c context.Context, r interface{}, sd upstream.ServerDiscovery) error {
			ss := sd.OnServers()
			if len(ss) == 1 {
				return ss[0].Serve(c, r)
			}
			return selectNextServer(ctx, ss).Serve(c, r)
		})
}

type weightedRRServerCtx struct {
	lock   sync.Mutex
	count  int
	caches map[string]*weightedRRServer
}

type weightedRRServer struct {
	CurrentWeight int
	upstream.Server
}

func selectNextServer(ctx *weightedRRServerCtx, ss upstream.Servers) upstream.Server {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	var total int
	var selected *weightedRRServer
	for i, _len := 0, len(ss); i < _len; i++ {
		weight := upstream.GetServerWeight(ss[i])
		total += weight

		id := ss[i].ID()
		ws, ok := ctx.caches[id]
		if ok {
			ws.CurrentWeight += weight
		} else {
			ws = &weightedRRServer{Server: ss[i], CurrentWeight: weight}
			ctx.caches[id] = ws
		}

		if selected == nil || selected.CurrentWeight < ws.CurrentWeight {
			selected = ws
		}
	}

	// We clean the down servers only each 1000 times.
	if ctx.count++; ctx.count >= 1000 {
		for id := range ctx.caches {
			if !ss.Contains(id) {
				delete(ctx.caches, id)
			}
		}
		ctx.count = 0
	}

	selected.CurrentWeight -= total
	return selected.Server
}
