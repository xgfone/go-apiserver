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
	"github.com/xgfone/go-apiserver/log"
)

// DefaultService is the default task service.
var DefaultService = NewService(context.Background())

// IsActivated is equal to DefaultService.IsActivated().
func IsActivated() bool { return DefaultService.IsActivated() }

// Context is equal to DefaultService.Context().
func Context() context.Context { return DefaultService.Context() }

// RunFunc is eqaul to DefaultService.RunFunc(f).
func RunFunc(f func(context.Context)) { DefaultService.RunFunc(f) }

// RunTask is equal to DefaultService.RunTask(taskID, taskName, taskFunc).
func RunTask(taskID, taskName string, taskFunc func(context.Context)) {
	DefaultService.RunTask(taskID, taskName, taskFunc)
}

type contextInfo struct {
	context.CancelFunc
	context.Context
}

// Service is a task-based service.
type Service struct {
	parent  context.Context
	context atomic.Value
	clock   sync.Mutex
	tasks   sync.Map
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
func (m *Service) Context() context.Context {
	return m.context.Load().(contextInfo).Context
}

// IsActivated reports whether the task service is activated.
func (m *Service) IsActivated() bool { return m.Context() != nil }

// Activate implements the interface service.Service.
func (m *Service) Activate() {
	if !m.IsActivated() {
		m.clock.Lock()
		if !m.IsActivated() {
			ctx, cancel := context.WithCancel(m.parent)
			m.context.Store(contextInfo{Context: ctx, CancelFunc: cancel})
		}
		m.clock.Unlock()
	}
}

// Deactivate implements the interface service.Service.
func (m *Service) Deactivate() {
	if m.IsActivated() {
		var cancel func()
		m.clock.Lock()
		if m.IsActivated() {
			cancel = m.context.Swap(contextInfo{}).(contextInfo).CancelFunc
		}
		m.clock.Unlock()
		if cancel != nil {
			cancel()
		}
	}
}

// RunFunc runs the function f only if the task service is activated.
func (m *Service) RunFunc(f func(context.Context)) {
	if ctx := m.Context(); ctx != nil {
		go f(ctx)
	}
}

// RunTask registers the task by taskID and runs the task asynchronously
// only if the task service is activated and the taskID is not registered.
func (m *Service) RunTask(taskID, taskName string, taskFunc func(context.Context)) {
	if taskID == "" || taskName == "" {
		panic("the task id or name must not be empty")
	}

	if ctx := m.Context(); ctx != nil {
		if value, loaded := m.tasks.LoadOrStore(taskID, taskName); !loaded {
			go m.run(ctx, taskID, taskFunc)
		} else if oldTask := value.(string); oldTask != taskName {
			log.Warn("another task is being run", "taskid", taskID,
				"oldtask", oldTask, "newtask", taskName)
		}
	}
}

func (m *Service) run(c context.Context, id string, f func(context.Context)) {
	defer m.tasks.Delete(id)
	f(c)
}
