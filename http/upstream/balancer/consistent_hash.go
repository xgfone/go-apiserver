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
// The policy name is "consistent_hash".
func ConsistentHash(callback SelectedServerCallback, hash func(*http.Request) int) Balancer {
	return NewForwarder("consistent_hash",
		func(w http.ResponseWriter, r *http.Request, s upstream.Servers) error {
			return serverCallback(callback, w, r, s[hash(r)%len(s)])
		})
}
