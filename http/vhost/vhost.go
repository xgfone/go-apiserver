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

// Package vhost provides the virtual host function.
package vhost

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/http/matcher"
	"github.com/xgfone/go-apiserver/tools/maps"
)

const (
	exactMatch = iota
	suffixMatch
)

type vhost struct {
	Handler http.Handler
	VHost   string

	Type int
	Prio int
}

func newVHost(host string, handler http.Handler) (vhost vhost) {
	if handler == nil {
		panic("vhost handler is nil")
	}

	if _len := len(host); _len == 0 {
		panic("vhost is empty")
	} else if _len > 2 && host[0] == '*' && host[1] == '.' { // Suffix Match
		vhost.Type = suffixMatch
		vhost.VHost = host[1:]
		vhost.Prio = 2
	} else { // Exact Match
		vhost.Type = exactMatch
		vhost.VHost = host
		vhost.Prio = 1
	}

	vhost.Handler = handler
	return
}

func (h vhost) MatchHost(host string) (ok bool) {
	switch h.Type {
	case exactMatch:
		ok = host == h.VHost

	case suffixMatch:
		ok = strings.HasSuffix(host, h.VHost)
	}

	return
}

type vhosts []vhost

func (hs vhosts) Len() int      { return len(hs) }
func (hs vhosts) Swap(i, j int) { hs[i], hs[j] = hs[j], hs[i] }
func (hs vhosts) Less(i, j int) bool {
	if hs[i].Prio < hs[j].Prio {
		return true
	} else if hs[i].Prio == hs[j].Prio && len(hs[i].VHost) > len(hs[j].VHost) {
		return true
	}
	return false
}

type defaultVHost struct{ http.Handler }
type vhostsWrapper struct{ vhosts }

// DefaultManager is the default global virtual host manager.
var DefaultManager = NewManager()

// Manager is used to manage a set of virtual hosts.
type Manager struct {
	// GetHostname is used to get the hostname from the request
	// to match the virtual hosts.
	//
	// Default: matcher.GetHost
	GetHostname func(*http.Request) (hostname string)

	// HandleHTTP is used to wrap the matched virtual host to decide
	// how to handle the request.
	//
	// Notice: If h is nil, it indicates that no virtual host matches the request.
	//
	// Default: h.ServeHTTP(w, r) or handler.Handler404.ServeHTTP(w, r)
	HandleHTTP func(w http.ResponseWriter, r *http.Request, h http.Handler)

	lock   sync.RWMutex
	vhosts map[string]vhost

	handler      atomic.Value
	defaultVHost atomic.Value
}

// NewManager returns a new virtual host manager.
func NewManager() *Manager {
	m := &Manager{vhosts: make(map[string]vhost, 8)}
	m.SetDefaultVHost(nil)
	m.updateVHosts()
	return m
}

func (m *Manager) getReqHost(r *http.Request) string {
	if m.GetHostname != nil {
		return m.GetHostname(r)
	}
	return matcher.GetHost(r)
}

func (m *Manager) updateVHosts() {
	vhosts := vhosts(maps.Values(m.vhosts))
	sort.Stable(vhosts)
	m.handler.Store(vhostsWrapper{vhosts: vhosts})
}

func (m *Manager) handlerHTTP(w http.ResponseWriter, r *http.Request, h http.Handler) {
	if m.HandleHTTP != nil {
		m.HandleHTTP(w, r, h)
	} else if h != nil {
		h.ServeHTTP(w, r)
	} else {
		handler.Handler404.ServeHTTP(w, r)
	}
}

// ServeHTTP implements the interface http.Handler.
func (m *Manager) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	vhosts := m.handler.Load().(vhostsWrapper).vhosts
	for i, _len := 0, len(vhosts); i < _len; i++ {
		if vhosts[i].MatchHost(m.getReqHost(r)) {
			m.handlerHTTP(rw, r, vhosts[i].Handler)
			return
		}
	}
	m.handlerHTTP(rw, r, m.GetDefaultVHost())
}

// GetDefaultVHost returns the default virtual host and handler.
func (m *Manager) GetDefaultVHost() http.Handler {
	return m.defaultVHost.Load().(defaultVHost).Handler
}

// SetDefaultVHost resets the default virtual host and handler.
func (m *Manager) SetDefaultVHost(handler http.Handler) {
	m.defaultVHost.Store(defaultVHost{Handler: handler})
}

// AddVHost adds the virtual host with the handler.
//
// The virtual host supports the exact or suffix match, like "www.example.com"
// or "*.example.com".
func (m *Manager) AddVHost(vhost string, handler http.Handler) (err error) {
	vh := newVHost(vhost, handler)

	m.lock.Lock()
	if maps.Add(m.vhosts, vhost, vh) {
		m.updateVHosts()
	} else {
		err = fmt.Errorf("vhost '%s' has been added", vhost)
	}
	m.lock.Unlock()

	return
}

// DelVHost deletes the given vritual host, and returns the handler.
//
// If the virtual host does not exist, reutrn nil.
func (m *Manager) DelVHost(vhost string) (handler http.Handler) {
	if vhost == "" {
		panic("vhost is empty")
	}

	m.lock.Lock()
	if vh, ok := maps.Pop(m.vhosts, vhost); ok {
		handler = vh.Handler
		m.updateVHosts()
	}
	m.lock.Unlock()

	return
}

// GetVHost returns the handler of the given virtual host.
//
// If the virtual host does not exist, reutrn nil.
func (m *Manager) GetVHost(vhost string) (handler http.Handler) {
	if vhost == "" {
		panic("vhost is empty")
	}

	m.lock.RLock()
	if vh, ok := m.vhosts[vhost]; ok {
		handler = vh.Handler
	}
	m.lock.RUnlock()
	return
}

// GetVHosts returns all the virtual hosts and handlers.
func (m *Manager) GetVHosts() (vhosts map[string]http.Handler) {
	m.lock.RLock()
	vhosts = make(map[string]http.Handler, len(m.vhosts))
	for host, vhost := range m.vhosts {
		vhosts[host] = vhost.Handler
	}
	m.lock.RUnlock()
	return
}
