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

package wait

import (
	"context"
	"math/rand"
	"time"
)

// Runner is a runner to run a task.
type Runner func(context.Context) (end bool, err error)

// ForeverRunner converts a function running forever to Runner.
func ForeverRunner(runner func(context.Context)) Runner {
	return func(ctx context.Context) (end bool, err error) {
		runner(ctx)
		return false, nil
	}
}

// Loop is an interface to loop until a condition reaches.
type Loop interface {
	// If f returns true or an error, it terminates looping.
	Run(context.Context, Runner) error
}

var _ Loop = JitterLoop{}

// JitterLoop loops running the task every jittered interval duration.
type JitterLoop struct {
	// The delay duration to run the task first.
	FirstDelay time.Duration

	// If true, immediately run the task first without waiting for the interval.
	FirstInstant bool

	// The interval duration between two runs of f.
	Interval time.Duration

	// If true, the interval is computed after f runs.
	// If false, the interval includes the runtime for f.
	Sliding bool

	// If positive, the Interval is jittered before every run of f.
	// If not positive, the Interval is unchanged and not jittered.
	JitterFactor float64
}

// NewJitterLoop is equal to NewJitterLoopWithFirstRun(true, 0, interval, sliding, jitterFactor).
func NewJitterLoop(interval time.Duration, sliding bool, jitterFactor float64) JitterLoop {
	return NewJitterLoopWithFirstRun(true, 0, interval, sliding, jitterFactor)
}

// NewJitterLoopWithFirstRun returns a new JitterLoop.
func NewJitterLoopWithFirstRun(firstInstant bool, firstDelay time.Duration,
	interval time.Duration, sliding bool, jitterFactor float64) JitterLoop {
	return JitterLoop{
		FirstDelay:   firstDelay,
		FirstInstant: firstInstant,
		JitterFactor: jitterFactor,
		Interval:     interval,
		Sliding:      sliding,
	}
}

// RunForever runs the function f forever until the context c is cancelled.
func (l JitterLoop) RunForever(c context.Context, r func(context.Context)) error {
	return l.Run(c, ForeverRunner(r))
}

// Run implements the interface Loop, which loops running f every interval
// duration until the context is done or f returns true or an error.
//
// Notice: f may not be invoked if context is already expired.
func (l JitterLoop) Run(c context.Context, r Runner) error {
	if l.Interval < 0 {
		panic("JitterLoop: the interval duration must not be negative")
	}

	if l.FirstDelay > 0 {
		t := time.NewTimer(l.FirstDelay)
		select {
		case <-t.C:
		case <-c.Done():
			t.Stop()
			return c.Err()
		}
	}

	if l.Interval == 0 {
		return l.r0(c, r)
	}

	var run bool
	var t *time.Timer
	for {
		select {
		case <-c.Done():
			return c.Err()
		default:
		}

		interval := l.Interval
		if l.JitterFactor > 0.0 {
			interval = Jitter(interval, l.JitterFactor)
		}

		if !run {
			run = true
			if !l.FirstInstant {
				t = time.NewTimer(interval)
				select {
				case <-t.C:
				case <-c.Done():
					t.Stop()
					return c.Err()
				}
			}
		}

		if !l.Sliding {
			if t == nil {
				t = time.NewTimer(interval)
			} else {
				t.Reset(interval)
			}
		}

		if ok, err := safeRun(c, r); ok || err != nil {
			return err
		}

		if l.Sliding {
			if t == nil {
				t = time.NewTimer(interval)
			} else {
				t.Reset(interval)
			}
		}

		select {
		case <-t.C:
		case <-c.Done():
			t.Stop()
			return c.Err()
		}
	}
}

func (l JitterLoop) r0(c context.Context, r Runner) error {
	for {
		select {
		case <-c.Done():
			return c.Err()
		default:
			if ok, err := safeRun(c, r); ok || err != nil {
				return err
			}
		}
	}
}

// Jitter returns a time.Duration between duration and duration + maxFactor *
// duration.
//
// This allows clients to avoid converging on periodic behavior.
// If maxFactor is 0.0, a suggested default value will be chosen.
func Jitter(duration time.Duration, maxFactor float64) time.Duration {
	if maxFactor <= 0.0 {
		maxFactor = 1.0
	}
	return duration + time.Duration(rand.Float64()*maxFactor*float64(duration))
}

// JitterUntil is a convenient function to launch the jitter loop.
func JitterUntil(c context.Context, interval time.Duration, sliding bool, jitter float64, r Runner) error {
	return NewJitterLoop(interval, sliding, jitter).Run(c, r)
}

// Until is a syntactic sugar on top of JitterUntil with zero jitter factor
// and with sliding = true.
func Until(c context.Context, interval time.Duration, r Runner) error {
	return JitterUntil(c, interval, true, 0.0, r)
}

// Until2 is the the same as Until, but use func(context.Context) instead
// to as the runner.
func Until2(c context.Context, interval time.Duration, f func(context.Context)) error {
	return Until(c, interval, func(ctx context.Context) (bool, error) {
		f(ctx)
		return false, nil
	})
}
