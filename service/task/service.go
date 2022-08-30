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

// Package task provides a task-based service.
package task

import (
	"context"
	"sync"

	"github.com/xgfone/go-apiserver/internal/atomic"
)

// DefaultService is the default task service.
var DefaultService = NewService(context.Background())

// IsActivated is equal to DefaultService.IsActivated().
func IsActivated() bool { return DefaultService.IsActivated() }

// Context is equal to DefaultService.Context().
func Context() context.Context { return DefaultService.Context() }

// Run is eqaul to DefaultService.Run(r).
func Run(r Runner) { DefaultService.Run(r) }

// RunFunc is eqaul to DefaultService.Run(f).
func RunFunc(f RunnerFunc) { DefaultService.Run(f) }

// AsyncRunFunc is eqaul to DefaultService.Run(r).
func AsyncRunFunc(f AsyncRunnerFunc) { DefaultService.Run(f) }

// WrappedRunnerFunc wraps the runner function and returns a new one
// that runs the wrapped runner function only when service is activated.
//
// If service is nil, use DefaultService instead.
func WrappedRunnerFunc(service *Service, f RunnerFunc) RunnerFunc {
	return func(ctx context.Context) {
		if (service == nil && DefaultService.IsActivated()) ||
			(service != nil && service.IsActivated()) {
			f(ctx)
		}
	}
}

// Runner is the task runner.
type Runner interface {
	Run(context.Context)
}

// AsyncRunner converts the task runner to an async task runner.
func AsyncRunner(r Runner) Runner { return AsyncRunnerFunc(r.Run) }

// RunnerFunc is the runner function.
type RunnerFunc func(context.Context)

// Run implements the interface Runner.
func (r RunnerFunc) Run(c context.Context) { r(c) }

// AsyncRunnerFunc is the async runner function.
type AsyncRunnerFunc func(context.Context)

// Run implements the interface Runner.
func (r AsyncRunnerFunc) Run(c context.Context) { go r(c) }

type contextInfo struct {
	context.CancelFunc
	context.Context
}

// Service is a service to run the task when it is activated.
type Service struct {
	parent  context.Context
	context atomic.Value
	clock   sync.Mutex
}

// NewService returns a new task service.
//
// If parent is nil, use context.Background() instead.
func NewService(parent context.Context) *Service {
	if parent == nil {
		parent = context.Background()
	}

	s := &Service{parent: parent}
	s.context.Store(contextInfo{})
	return s
}

// Context returns the inner context, which will be cancelled
// when the task service is deactivated.
//
// Return nil instead if the task service is not activated.
func (s *Service) Context() context.Context {
	return s.context.Load().(contextInfo).Context
}

// IsActivated reports whether the task service is activated.
func (s *Service) IsActivated() bool { return s.Context() != nil }

// Activate implements the interface service.Service.
func (s *Service) Activate() {
	if !s.IsActivated() {
		s.clock.Lock()
		if !s.IsActivated() {
			ctx, cancel := context.WithCancel(s.parent)
			s.context.Store(contextInfo{Context: ctx, CancelFunc: cancel})
		}
		s.clock.Unlock()
	}
}

// Deactivate implements the interface service.Service.
func (s *Service) Deactivate() {
	if s.IsActivated() {
		var cancel func()
		s.clock.Lock()
		if s.IsActivated() {
			cancel = s.context.Swap(contextInfo{}).(contextInfo).CancelFunc
		}
		s.clock.Unlock()
		if cancel != nil {
			cancel()
		}
	}
}

// Run runs the runner r only if the task service is activated.
func (s *Service) Run(r Runner) {
	if ctx := s.Context(); ctx != nil {
		r.Run(ctx)
	}
}
