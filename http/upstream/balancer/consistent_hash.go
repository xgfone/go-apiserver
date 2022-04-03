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
	"net/http"

	"github.com/xgfone/go-apiserver/http/upstream"
)

// ConsistentHash returns a new balancer based on the consistent hash.
//
// The policy name is equal to hashPolicy with the prefix "consistent_hash_".
func ConsistentHash(hashPolicy string, hash func(*http.Request) int) Balancer {
	return NewBalancer("consistent_hash_"+hashPolicy,
		func(w http.ResponseWriter, r *http.Request, ss upstream.Servers) error {
			_len := len(ss)
			if _len == 1 {
				return ss[0].HandleHTTP(w, r)
			}
			return ss[hash(r)%_len].HandleHTTP(w, r)
		})
}
