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
	"time"

	up "github.com/xgfone/go-apiserver/http/upstream"
)

// Retry returns a new balancer to retry the rest servers when failing to
// forward the request.
//
// Notice: It will retry the same upstream server for the sourceip or consistent
// hash balancer.
func Retry(balancer Balancer, interval time.Duration) Balancer {
	if balancer == nil {
		panic("RetryBalancer: the wrapped balancer is nil")
	}
	return retry{Balancer: balancer, interval: interval}
}

type retry struct {
	interval time.Duration
	Balancer
}

func (b retry) WrappedBalancer() Balancer { return b.Balancer }
func (b retry) Forward(w http.ResponseWriter, r *http.Request, f func() up.Servers) (err error) {
	ss := f()
	_len := len(ss)
	if _len == 1 {
		return ss[0].HandleHTTP(w, r)
	}

	c := r.Context()
	for ; _len > 0; _len-- {
		select {
		case <-c.Done():
			return c.Err()
		default:
		}

		if err = b.Balancer.Forward(w, r, f); err == nil {
			break
		}

		if b.interval > 0 {
			timer := time.NewTimer(b.interval)
			select {
			case <-timer.C:
			case <-c.Done():
				timer.Stop()
				return c.Err()
			}
		}
	}

	return
}
