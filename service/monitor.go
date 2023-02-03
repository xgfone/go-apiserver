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

package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xgfone/go-apiserver/log"
	atomic2 "github.com/xgfone/go-apiserver/tools/atomic"
)

// DefaultCheckConfig is the global default check configuration.
var DefaultCheckConfig = CheckConfig{
	Interval: time.Second * 2,
	Timeout:  time.Second,
	Failure:  1,
}

// CheckConfig is used to check the state of the checker.
type CheckConfig struct {
	Interval time.Duration
	Timeout  time.Duration
	Failure  int
}

// Checker is used to check whether a condition is ok.
type Checker interface {
	Check(context.Context) (ok bool, err error)
}

// CheckerFunc is the checker function.
type CheckerFunc func(context.Context) (ok bool, err error)

// Check implements the interface Checker.
func (f CheckerFunc) Check(c context.Context) (ok bool, err error) { return f(c) }

// NothingChecker returns a new nothing checker that always returns (false, nil).
func NothingChecker() Checker {
	return CheckerFunc(func(context.Context) (ok bool, e error) { return })
}

// Monitor is used to control that a group of services activate or deactivate.
type Monitor struct {
	clock  sync.RWMutex
	cancel context.CancelFunc

	state   int32
	failure int
	cconfig atomic.Value
	checker atomic2.Value
	service atomic2.Value
	setsvc  chan struct{}
}

// NewMonitor returns a new service monitor.
//
// If cconf is nil, use DefaultCheckConfig instead.
func NewMonitor(svc Service, checker Checker, cconf *CheckConfig) *Monitor {
	config := DefaultCheckConfig
	if cconf != nil {
		config = *cconf
	}
	if config.Interval < 1 {
		panic("check interval must be greater than 0")
	}

	if svc == nil {
		panic("the service must not be nil")
	}
	if checker == nil {
		panic("the checker must not be nil")
	}

	m := new(Monitor)
	m.SetService(svc)
	m.SetChecker(checker)
	m.SetCheckConfig(config)
	m.setsvc = make(chan struct{}, 1)
	return m
}

// GetCheckConfig returns the stored check config.
func (m *Monitor) GetCheckConfig() CheckConfig {
	return m.cconfig.Load().(CheckConfig)
}

// SetCheckConfig resets the check service.
func (m *Monitor) SetCheckConfig(c CheckConfig) {
	if c.Interval < 1 {
		panic("service.Monitor: the check interval must be greater than 0")
	}
	m.cconfig.Store(c)
}

// GetChecker returns the stored checker.
func (m *Monitor) GetChecker() Checker {
	return m.checker.Load().(Checker)
}

// SetChecker resets the checker.
func (m *Monitor) SetChecker(c Checker) {
	if c == nil {
		panic("service.Monitor: the checker must not be nil")
	}
	m.checker.Store(c)
}

// GetService returns the stored service.
func (m *Monitor) GetService() Service {
	return m.service.Load().(Service)
}

// SetService resets the service.
func (m *Monitor) SetService(s Service) {
	if s == nil {
		panic("service.Monitor: the service must not be nil")
	}
	m.service.Store(s)
	select {
	case m.setsvc <- struct{}{}:
	default:
	}
}

// IsActive reports whether the monitor is active, that's,
// the service associated with the monitor is activated.
func (m *Monitor) IsActive() bool {
	return atomic.LoadInt32(&m.state) == 1
}

// IsActivated reports whether the monitor is activated,
// that's, the monitor is enabled or running.
func (m *Monitor) IsActivated() bool {
	m.clock.RLock()
	ok := m.cancel != nil
	m.clock.RUnlock()
	return ok
}

// Activate implements the interface Service.
func (m *Monitor) Activate() {
	m.clock.Lock()
	if m.cancel == nil {
		var ctx context.Context
		ctx, m.cancel = context.WithCancel(context.Background())
		go m.run(ctx)
	}
	m.clock.Unlock()
}

// Deactivate implements the interface Service.
func (m *Monitor) Deactivate() {
	m.clock.Lock()
	defer m.clock.Unlock()

	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
		m.deactivate()
	}
}

func (m *Monitor) run(ctx context.Context) {
	conf := m.GetCheckConfig()
	m.check(ctx, conf)

	ticker := time.NewTicker(conf.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-m.setsvc:
			if m.IsActive() {
				safeRun(m.GetService().Activate)
			} else {
				safeRun(m.GetService().Deactivate)
			}

		case <-ticker.C:
			newconf := m.GetCheckConfig()
			if newconf.Interval != conf.Interval {
				ticker.Reset(newconf.Interval)
				conf.Interval = newconf.Interval
			}
			m.check(ctx, newconf)
		}
	}
}

func safeRun(f func()) {
	defer log.WrapPanic()
	f()
}

func (m *Monitor) check(ctx context.Context, cconfig CheckConfig) {
	defer log.WrapPanic()

	if cconfig.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cconfig.Timeout)
		defer cancel()
	}

	checker := m.GetChecker()
	ok, err := checker.Check(ctx)
	if err != nil {
		if m.failure++; m.failure > cconfig.Failure {
			m.deactivate()
		}

		if named, ok := checker.(interface{ Name() string }); ok {
			log.Error("checker failed", "checker", named.Name(), "err", err)
		} else {
			log.Ef(err, "checker failed")
		}
	} else {
		if m.failure > 0 {
			m.failure = 0
		}

		if ok {
			m.activate()
		} else {
			m.deactivate()
		}
	}
}

func (m *Monitor) activate() {
	if atomic.CompareAndSwapInt32(&m.state, 0, 1) {
		m.GetService().Activate()
	}
}

func (m *Monitor) deactivate() {
	if atomic.CompareAndSwapInt32(&m.state, 1, 0) {
		m.GetService().Deactivate()
	}
}
