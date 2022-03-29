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

package servicemonitor

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/service"
)

// DefaultCheckConfig is the global default check configuration.
var DefaultCheckConfig = CheckConfig{
	Tick:    time.Second * 2,
	Timeout: time.Second,
	Failure: 1,
}

// CheckConfig is used to check the state of the checker.
type CheckConfig struct {
	Tick    time.Duration
	Timeout time.Duration
	Failure int
}

// Checker is used to check whether a condition is ok.
type Checker interface {
	Check(context.Context) (ok bool, err error)
}

// CheckerFunc is the checker function.
type CheckerFunc func(context.Context) (ok bool, err error)

// Check implements the interface Checker.
func (f CheckerFunc) Check(c context.Context) (ok bool, err error) { return f(c) }

// Monitor is used to control that a group of services activate or deactivate.
type Monitor struct {
	clock  sync.RWMutex
	cancel context.CancelFunc

	state   int32
	failure int
	cconfig CheckConfig
	checker Checker
	service service.Service
}

// NewMonitor returns a new service monitor.
//
// If cconf is nil, use DefaultCheckConfig instead.
func NewMonitor(svc service.Service, checker Checker, cconf *CheckConfig) *Monitor {
	config := DefaultCheckConfig
	if cconf != nil {
		config = *cconf
	}
	if config.Tick < 1 {
		panic("checker tick must be greater than 0")
	}

	if svc == nil {
		panic("the service must not be nil")
	}
	if checker == nil {
		panic("the checker must not be nil")
	}

	return &Monitor{service: svc, checker: checker, cconfig: config}
}

// IsActive reports whether the monitor is active,
// that's, the service is activated.
func (c *Monitor) IsActive() bool {
	return atomic.LoadInt32(&c.state) == 1
}

// IsActivated reports whether the monitor is activated,
// that's, the monitor is enabled or running.
func (c *Monitor) IsActivated() bool {
	c.clock.RLock()
	ok := c.cancel != nil
	c.clock.RUnlock()
	return ok
}

// Activate implements the interface Service.
func (c *Monitor) Activate() {
	c.clock.Lock()
	if c.cancel == nil {
		var ctx context.Context
		ctx, c.cancel = context.WithCancel(context.Background())
		go c.run(ctx)
	}
	c.clock.Unlock()
}

// Deactivate implements the interface Service.
func (c *Monitor) Deactivate() {
	c.clock.Lock()
	defer c.clock.Unlock()

	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
		c.deactivate()
	}
}

func (c *Monitor) run(ctx context.Context) {
	c.check(ctx)

	ticker := time.NewTicker(c.cconfig.Tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			c.check(ctx)
		}
	}
}

func (c *Monitor) check(ctx context.Context) {
	defer log.WrapPanic()

	if c.cconfig.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.cconfig.Timeout)
		defer cancel()
	}

	ok, err := c.checker.Check(ctx)
	if err != nil {
		if c.failure++; c.failure > c.cconfig.Failure {
			c.deactivate()
		}

		if named, ok := c.checker.(interface{ Name() string }); ok {
			log.Error("checker failed", "checker", named.Name(), "err", err)
		} else {
			log.Ef(err, "checker failed")
		}
	} else {
		if c.failure > 0 {
			c.failure = 0
		}

		if ok {
			c.activate()
		} else {
			c.deactivate()
		}
	}
}

func (c *Monitor) activate() {
	if atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		c.service.Activate()
	}
}

func (c *Monitor) deactivate() {
	if atomic.CompareAndSwapInt32(&c.state, 1, 0) {
		c.service.Deactivate()
	}
}
