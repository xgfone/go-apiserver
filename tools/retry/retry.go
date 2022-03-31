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

// Package retry implements some retry policies to call a function.
package retry

import (
	"context"
	"time"
)

// Retry is an interface to retry the function until a condition reaches.
type Retry interface {
	// If f returns true or nil, it terminates retry.
	Run(c context.Context, f func(context.Context) (success bool, err error)) error
}

// NewPeriodicLoopRetry returns a new retry to call a runner function
// periodically until the context is done or it reaches the number.
//
// number is the times to retry to call a function, which must be positive.
//
// interval is the interval duration between two callings.
// If 0, it immediately retries to call.
func NewPeriodicLoopRetry(number int, interval time.Duration) Retry {
	return periodicLoopRetry{Number: number, Interval: interval}
}

type periodicLoopRetry struct {
	Interval time.Duration
	Number   int
}

func (r periodicLoopRetry) Run(c context.Context, f func(context.Context) (bool, error)) error {
	if r.Number < 1 {
		panic("the retry number must be positive")
	}

	var ok bool
	var err error
	for n := r.Number; n > 0; n-- {
		select {
		case <-c.Done():
			return c.Err()
		default:
		}

		if ok, err = f(c); ok || err == nil {
			return err
		}

		if r.Interval > 0 {
			t := time.NewTimer(r.Interval)
			select {
			case <-t.C:
			case <-c.Done():
				t.Stop()
				return c.Err()
			}
		}
	}

	return err
}
