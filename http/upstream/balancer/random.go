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
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/xgfone/go-apiserver/http/upstream"
)

func newRandom() func(int) int {
	lock := new(sync.Mutex)
	random := rand.New(rand.NewSource(time.Now().UnixNano())).Intn
	return func(i int) (n int) {
		lock.Lock()
		n = random(i)
		lock.Unlock()
		return
	}
}

func init() {
	registerBuiltinBuidler("random", Random)
	registerBuiltinBuidler("weight_random", WeightedRandom)
}

// Random returns a new balancer based on the random.
//
// The policy name is "random".
func Random() Balancer {
	random := newRandom()
	return NewBalancer("random",
		func(w http.ResponseWriter, r *http.Request, f func() upstream.Servers) error {
			ss := f()
			_len := len(ss)
			if _len == 1 {
				return ss[0].HandleHTTP(w, r)
			}
			return ss[random(_len)].HandleHTTP(w, r)
		})
}

// WeightedRandom returns a new balancer based on the roundrobin and weight.
//
// The policy name is "weight_random".
func WeightedRandom() Balancer {
	random := newRandom()
	return NewBalancer("weight_random",
		func(w http.ResponseWriter, r *http.Request, f func() upstream.Servers) error {
			ss := f()
			_len := len(ss)
			if _len == 1 {
				return ss[0].HandleHTTP(w, r)
			}

			var total int
			for i := 0; i < _len; i++ {
				total += upstream.GetServerWeight(ss[i])
			}

			pos := random(total)
			for {
				var total int
				for i := 0; i < _len; i++ {
					total += upstream.GetServerWeight(ss[i])
					if pos <= total {
						return ss[i].HandleHTTP(w, r)
					}
				}
				pos %= total
			}
		})
}
