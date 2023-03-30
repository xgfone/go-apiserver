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

	"github.com/xgfone/go-apiserver/upstream"
	"github.com/xgfone/go-generics/maps"
)

var _ upstream.ServerWrapper = new(upserver)

type upserver struct {
	upstream.Server
	status atomic.Value
}

func newUpServer(server upstream.Server) *upserver {
	s := &upserver{Server: server}
	s.status.Store(upstream.ServerStatusOnline)
	return s
}

func (s *upserver) Unwrap() upstream.Server {
	return s.Server
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

	maps.Clear(m.servers)
	maps.AddSlice(m.servers, servers, func(s upstream.Server) (string, *upserver) {
		return s.ID(), newUpServer(s)
	})
	m.updateServers()
}

func (m *serversManager) UpsertServers(servers ...upstream.Server) {
	m.slock.Lock()
	defer m.slock.Unlock()

	maps.AddSlice(m.servers, servers, func(s upstream.Server) (string, *upserver) {
		return s.ID(), newUpServer(s)
	})
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
	onservers := upstream.AcquireServers(len(m.servers))
	allservers := upstream.AcquireServers(len(m.servers))
	offservers := upstream.AcquireServers(0)
	for _, server := range m.servers {
		allservers = append(allservers, server)
		switch server.Status() {
		case upstream.ServerStatusOnline:
			onservers = append(onservers, server.Server)
		case upstream.ServerStatusOffline:
			offservers = append(offservers, server.Server)
		}
	}

	// For online
	if len(onservers) == 0 {
		oldservers := m.onservers.Swap(upstream.Servers{}).(upstream.Servers)
		upstream.ReleaseServers(oldservers)
		upstream.ReleaseServers(onservers)
	} else {
		sort.Stable(onservers)
		oldservers := m.onservers.Swap(onservers).(upstream.Servers)
		upstream.ReleaseServers(oldservers)
	}

	// For offline
	if len(offservers) == 0 {
		oldservers := m.offservers.Swap(upstream.Servers{}).(upstream.Servers)
		upstream.ReleaseServers(oldservers)
		upstream.ReleaseServers(offservers)
	} else {
		sort.Stable(offservers)
		oldservers := m.offservers.Swap(offservers).(upstream.Servers)
		upstream.ReleaseServers(oldservers)
	}

	// For all
	if len(allservers) == 0 {
		oldservers := m.allservers.Swap(upstream.Servers{}).(upstream.Servers)
		upstream.ReleaseServers(oldservers)
		upstream.ReleaseServers(allservers)
	} else {
		sort.Stable(allservers)
		oldservers := m.allservers.Swap(allservers).(upstream.Servers)
		upstream.ReleaseServers(oldservers)
	}
}
