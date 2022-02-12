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
	"time"

	"github.com/xgfone/go-apiserver/http/upstream"
)

func init() {
	registerBuiltinBuidler("random", Random)
	registerBuiltinBuidler("weight_random", WeightedRandom)
}

// Random returns a new balancer based on the random.
//
// The policy name is "random".
func Random() Balancer {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	return NewBalancer("random",
		func(w http.ResponseWriter, r *http.Request, s upstream.Servers) error {
			return forward(w, r, s[random.Intn(len(s))])
		})
}

// WeightedRandom returns a new balancer based on the roundrobin and weight.
//
// The policy name is "weight_random".
func WeightedRandom() Balancer {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	return NewBalancer("weight_random",
		func(w http.ResponseWriter, r *http.Request, s upstream.Servers) error {
			pos := uint64(random.Intn(len(s)))
			return forward(w, r, calcServerOnWeight(s, pos))
		})
}
