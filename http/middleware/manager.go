// Copyright 2023 xgfone
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

package middleware

import "net/http"

// Manager is used to manage a set of the middlewares.
type Manager struct {
	origal  http.Handler
	handler http.Handler
	mdws    Middlewares
}

// NewManager returns a new middleware manager.
func NewManager(handler http.Handler) *Manager {
	return &Manager{origal: handler, handler: handler}
}

func (m *Manager) update() { m.handler = m.Handler(m.origal) }

// Middlewares returns the added middlewares.
func (m *Manager) Middlewares() Middlewares { return m.mdws }

// Handler wraps the http handler with the added middlewares,
// and return a new http handler.
func (m *Manager) Handler(handler http.Handler) http.Handler {
	return m.mdws.Handler(handler)
}

// SetHandler resets the http handler.
func (m *Manager) SetHandler(handler http.Handler) {
	m.origal = handler
	m.update()
}

// InsertFunc inserts the new function middlewares to the front.
func (m *Manager) InsertFunc(ms ...MiddlewareFunc) {
	if len(ms) == 0 {
		return
	}

	m.mdws = m.mdws.InsertFunc(ms...)
	m.update()
}

// AppendFunc appends the new function middlewares.
func (m *Manager) AppendFunc(ms ...MiddlewareFunc) {
	if len(ms) == 0 {
		return
	}

	m.mdws.AppendFunc(ms...)
	m.update()
}

// Insert inserts the new middlewares to the front.
func (m *Manager) Insert(ms ...Middleware) {
	if len(ms) == 0 {
		return
	}

	mdws := make(Middlewares, 0, len(m.mdws)+len(ms))
	mdws = append(mdws, ms...)
	mdws = append(mdws, m.mdws...)
	m.mdws = mdws
	m.update()
}

// Append appends the new middlewares.
func (m *Manager) Append(ms ...Middleware) {
	if len(ms) == 0 {
		return
	}

	m.mdws = m.mdws.Append(ms...)
	m.update()
}

// Reset resets the middlewares to ms.
func (m *Manager) Reset(ms ...Middleware) {
	m.mdws = append(Middlewares{}, ms...)
	m.update()
}

// ServeHTTP implements the interface http.Handler.
func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.handler.ServeHTTP(w, r)
}

var _ http.Handler = new(Manager)
