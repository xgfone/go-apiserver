// Copyright 2022~2023 xgfone
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
	"time"

	"github.com/xgfone/go-apiserver/upstream"
)

var _ Balancer = Retry{}

// Retry is used to retry the rest servers when failing to forward the request.
//
// Notice: It will retry the same upstream server for the sourceip or consistent
// hash balancer.
type Retry struct {
	Interval time.Duration
	Balancer
}

// NewRetry returns a new retry balancer.
func NewRetry(balancer Balancer, interval time.Duration) Retry {
	return Retry{Balancer: balancer, Interval: interval}
}

// Forward overrides the Forward method.
func (b Retry) Forward(c context.Context, r interface{}, sd upstream.ServerDiscovery) (err error) {
	ss := sd.OnServers()
	_len := len(ss)
	if _len == 1 {
		return ss[0].Serve(c, r)
	}

	for ; _len > 0; _len-- {
		select {
		case <-c.Done():
			return c.Err()
		default:
		}

		if err = b.Balancer.Forward(c, r, sd.OnServers()); err == nil {
			break
		}

		if b.Interval > 0 {
			timer := time.NewTimer(b.Interval)
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
