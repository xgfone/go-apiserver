// Copyright 2022 xgfone
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

	up "github.com/xgfone/go-apiserver/http/upstream"
)

// Retry returns a new balancer to retry the rest servers when failing to
// forward the request.
func Retry(balancer Balancer) Balancer {
	if balancer == nil {
		panic("RetryBalancer: the wrapped balancer is nil")
	}
	return retry{Balancer: balancer}
}

type retry struct{ Balancer }

func (f retry) WrappedBalancer() Balancer { return f.Balancer }
func (f retry) Forward(w http.ResponseWriter, r *http.Request, s up.Servers) (err error) {
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
