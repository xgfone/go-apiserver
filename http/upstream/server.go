// Copyright 2021 xgfone
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

package upstream

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/xgfone/go-apiserver/nets"
)

// ErrNoAvailableServers is used to represents no available servers.
var ErrNoAvailableServers = errors.New("no available servers")

// Pre-define some server statuses.
const (
	ServerStatusOnline  ServerStatus = "on"
	ServerStatusOffline ServerStatus = "off"
)

// ServerStatus represents the server status.
type ServerStatus string

// IsOnline reports whether the server status is online.
func (s ServerStatus) IsOnline() bool { return s == ServerStatusOnline }

// IsOffline reports whether the server status is offline.
func (s ServerStatus) IsOffline() bool { return s == ServerStatusOffline }

// Server represents an upstream http server.
type Server interface {
	ID() string
	URL() URL
	Status() ServerStatus
	RuntimeState() nets.RuntimeState
	HandleHTTP(http.ResponseWriter, *http.Request) error
	Check(ctx context.Context, healthURL URL) error
}

// WeightedServer represents an upstream http server with the weight.
type WeightedServer interface {
	// Weight returns the weight of the server, which must be a positive integer.
	//
	// The bigger the value, the higher the weight.
	Weight() int

	Server
}

// ServerDiscovery is used to discover the servers.
type ServerDiscovery interface {
	OnlineNum() int
	OnServers() Servers
	OffServers() Servers
	AllServers() Servers
}

// Servers represents a group of the servers.
type Servers []Server

// OnlineNum implements the interface ServerDiscovery#OnlineNum
// to return the number of all the online servers.
func (ss Servers) OnlineNum() int { return len(ss.OnServers()) }

// OnServers implements the interface ServerDiscovery#OnServers
// to return all the online servers.
func (ss Servers) OnServers() Servers {
	var offline bool
	for _, s := range ss {
		if s.Status().IsOffline() {
			offline = true
			break
		}
	}
	if !offline {
		return ss
	}

	servers := make(Servers, 0, len(ss))
	for _, s := range ss {
		if s.Status().IsOnline() {
			servers = append(servers, s)
		}
	}
	return servers
}

// OffServers implements the interface ServerDiscovery#OffServers
// to return all the offline servers.
func (ss Servers) OffServers() Servers {
	var online bool
	for _, s := range ss {
		if s.Status().IsOnline() {
			online = true
			break
		}
	}
	if !online {
		return ss
	}

	servers := make(Servers, 0, len(ss))
	for _, s := range ss {
		if s.Status().IsOffline() {
			servers = append(servers, s)
		}
	}
	return servers
}

// AllServers implements the interface ServerDiscovery#AllServers
// to return all the servers.
func (ss Servers) AllServers() Servers { return ss }

// Contains reports whether the servers contains the server indicated by the id.
func (ss Servers) Contains(serverID string) bool {
	for _, s := range ss {
		if s.ID() == serverID {
			return true
		}
	}
	return false
}

// Sort the servers by the ASC order.
func (ss Servers) Len() int      { return len(ss) }
func (ss Servers) Swap(i, j int) { ss[i], ss[j] = ss[j], ss[i] }
func (ss Servers) Less(i, j int) bool {
	iw, jw := GetServerWeight(ss[i]), GetServerWeight(ss[j])
	if iw < jw {
		return true
	} else if iw == jw {
		return ss[i].ID() < ss[j].ID()
	} else {
		return false
	}
}

// GetServerWeight returns the weight of the server if it has implements
// the interface WeightedServer. Or return 1 instead.
func GetServerWeight(server Server) int {
	if ws, ok := server.(WeightedServer); ok {
		return ws.Weight()
	}
	return 1
}

// DefaultServersPool is the default servers pool.
var DefaultServersPool = NewServersPool(16)

// ServersPool is used to allocate and recycle the server slice.
type ServersPool struct{ pool sync.Pool }

// NewServersPool returns a new servers pool.
func NewServersPool(defaultCap int) *ServersPool {
	sp := &ServersPool{}
	sp.pool.New = func() interface{} { return make(Servers, 0, defaultCap) }
	return sp
}

// Acquire returns a server slice from the servers pool.
func (sp *ServersPool) Acquire() Servers { return sp.pool.Get().(Servers) }

// Release releases the servers into the pool.
func (sp *ServersPool) Release(servers Servers) {
	if cap(servers) > 0 {
		sp.pool.Put(servers[:0])
	}
}
