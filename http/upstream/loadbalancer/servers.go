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

package loadbalancer

import (
	"sort"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/http/upstream"
)

type upserver struct {
	upstream.Server
	status atomic.Value
}

func newUpServer(server upstream.Server) *upserver {
	s := &upserver{Server: server}
	s.status.Store(upstream.ServerStatusOnline)
	return s
}

func (s *upserver) Status() upstream.ServerStatus {
	return s.status.Load().(upstream.ServerStatus)
}

func (s *upserver) SetStatus(status upstream.ServerStatus) (ok bool) {
	return s.status.CompareAndSwap(s.Status(), status)
}

type serversManager struct {
	slock   sync.RWMutex
	servers map[string]*upserver

	onservers  atomic.Value
	offservers atomic.Value
	allservers atomic.Value
}

func newServersManager() *serversManager {
	m := &serversManager{servers: make(map[string]*upserver, 8)}
	m.allservers.Store(upstream.Servers{})
	m.offservers.Store(upstream.Servers{})
	m.onservers.Store(upstream.Servers{})
	return m
}

// OnlineNum implements the interface upstream.ServerDiscovery#OnlineNum.
func (m *serversManager) OnlineNum() int {
	return len(m.OnServers())
}

// OnServers implements the interface upstream.ServerDiscovery#OnServers.
func (m *serversManager) OnServers() upstream.Servers {
	return m.onservers.Load().(upstream.Servers)
}

// OffServers implements the interface upstream.ServerDiscovery#OffServers.
func (m *serversManager) OffServers() upstream.Servers {
	return m.offservers.Load().(upstream.Servers)
}

// AllServers implements the interface upstream.ServerDiscovery#AllServers.
func (m *serversManager) AllServers() upstream.Servers {
	return m.allservers.Load().(upstream.Servers)
}

func (m *serversManager) SetServerStatus(serverID string, status upstream.ServerStatus) {
	m.slock.RLock()
	if upserver, ok := m.servers[serverID]; ok {
		if upserver.SetStatus(status) {
			m.updateServers()
		}
	}
	m.slock.RUnlock()
}

func (m *serversManager) SetServerStatuses(statuses map[string]upstream.ServerStatus) {
	m.slock.RLock()
	var changed bool
	for serverID, status := range statuses {
		if upserver, ok := m.servers[serverID]; ok {
			if upserver.SetStatus(status) && !changed {
				changed = true
			}
		}
	}
	if changed {
		m.updateServers()
	}
	m.slock.RUnlock()
	return
}

func (m *serversManager) GetServer(serverID string) (server upstream.Server, ok bool) {
	m.slock.RLock()
	server, ok = m.servers[serverID]
	m.slock.RUnlock()
	return
}

func (m *serversManager) ResetServers(servers ...upstream.Server) {
	m.slock.Lock()
	defer m.slock.Unlock()

	allservers := make(map[string]*upserver, len(servers))
	for i, _len := 0, len(servers); i < _len; i++ {
		server := servers[i]
		allservers[server.ID()] = newUpServer(server)
	}

	m.servers = allservers
	m.updateServers()
}

func (m *serversManager) UpsertServers(servers ...upstream.Server) {
	m.slock.Lock()
	defer m.slock.Unlock()

	for i, _len := 0, len(servers); i < _len; i++ {
		server := servers[_len]
		m.servers[server.ID()] = newUpServer(server)
	}
	m.updateServers()
}

func (m *serversManager) RemoveServer(id string) {
	m.slock.Lock()
	defer m.slock.Unlock()

	if _, ok := m.servers[id]; ok {
		delete(m.servers, id)
		m.updateServers()
	}

	return
}

func (m *serversManager) updateServers() {
	onservers := upstream.DefaultServersPool.Acquire()
	offservers := upstream.DefaultServersPool.Acquire()
	allservers := upstream.DefaultServersPool.Acquire()
	for _, server := range m.servers {
		allservers = append(allservers, server)
		switch server.Status() {
		case upstream.ServerStatusOnline:
			onservers = append(onservers, server)
		case upstream.ServerStatusOffline:
			offservers = append(offservers, server)
		}
	}

	// For online
	if len(onservers) == 0 {
		oldservers := m.onservers.Swap(upstream.Servers{}).(upstream.Servers)
		upstream.DefaultServersPool.Release(oldservers)
		upstream.DefaultServersPool.Release(onservers)
	} else {
		sort.Stable(onservers)
		oldservers := m.onservers.Swap(onservers).(upstream.Servers)
		upstream.DefaultServersPool.Release(oldservers)
	}

	// For offline
	if len(offservers) == 0 {
		oldservers := m.offservers.Swap(upstream.Servers{}).(upstream.Servers)
		upstream.DefaultServersPool.Release(oldservers)
		upstream.DefaultServersPool.Release(offservers)
	} else {
		sort.Stable(offservers)
		oldservers := m.offservers.Swap(offservers).(upstream.Servers)
		upstream.DefaultServersPool.Release(oldservers)
	}

	// For all
	if len(allservers) == 0 {
		oldservers := m.allservers.Swap(upstream.Servers{}).(upstream.Servers)
		upstream.DefaultServersPool.Release(oldservers)
		upstream.DefaultServersPool.Release(allservers)
	} else {
		sort.Stable(allservers)
		oldservers := m.allservers.Swap(allservers).(upstream.Servers)
		upstream.DefaultServersPool.Release(oldservers)
	}
}
