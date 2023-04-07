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

	"github.com/xgfone/go-atomicvalue"
	"github.com/xgfone/go-checker"
)

// Monitor is used to control that a group of services activate or deactivate.
type Monitor struct {
	clock  sync.RWMutex
	cancel context.CancelFunc

	checker *checker.Checker
	service atomicvalue.Value[Service]
}

// NewMonitor returns a new service monitor with checker.DefaultConfig.
//
// If cond is nil use checker.DefaultCondition instead.
func NewMonitor(svc Service, cond checker.Condition) *Monitor {
	m := new(Monitor)
	m.checker = checker.NewChecker("servicemonitor", cond, m.activate)
	m.SetService(svc)
	return m
}

// GetCheckerConfig returns the stored checker config.
func (m *Monitor) GetCheckerConfig() checker.Config { return m.checker.Config() }

// SetCheckerConfig resets the checker config.
func (m *Monitor) SetCheckerConfig(config checker.Config) { m.checker.SetConfig(config) }

// GetChecker returns the stored checker.
func (m *Monitor) GetChecker() checker.Condition { return m.checker.Condition() }

// SetChecker resets the checker.
func (m *Monitor) SetChecker(c checker.Condition) {
	if c == nil {
		panic("service.Monitor: the checker must not be nil")
	}
	m.checker.SetCondition(c)
}

// GetService returns the stored service.
func (m *Monitor) GetService() Service { return m.service.Load() }

// SetService resets the service.
func (m *Monitor) SetService(s Service) {
	if s == nil {
		panic("service.Monitor: the service must not be nil")
	}
	m.service.Store(s)
	m.activate("", m.IsActive())
}

// IsActive reports whether the monitor is active, that's,
// the service associated with the monitor is activated.
func (m *Monitor) IsActive() bool { return m.checker.Ok() }

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
		go m.checker.Start(ctx)
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
		m.checker.SetOk(false)
	}
}

func (m *Monitor) activate(_ string, ok bool) {
	if ok {
		m.GetService().Activate()
	} else {
		m.GetService().Deactivate()
	}
}
