// Copyright 2021~2023 xgfone
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

// Package upstream provides some common upstream functions.
package upstream

import (
	"context"
	"errors"
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
	// Static information
	ID() string
	Type() string
	Info() interface{}

	// Dynamic information
	Status() ServerStatus
	RuntimeState() nets.RuntimeState

	// Handler
	Update(info interface{}) error
	Serve(ctx context.Context, req interface{}) error
	Check(context.Context) error
}

// ServerWrapper is a wrapper to wrap the upstream server.
type ServerWrapper interface {
	Unwrap() Server
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

// UnwrapServer unwraps the innest upsteam server.
func UnwrapServer(server Server) Server {
	if w, ok := server.(ServerWrapper); ok {
		UnwrapServer(w.Unwrap())
	}
	return server
}

// GetServerWeight returns the weight of the server if it has implements
// the interface WeightedServer. Or, check whether it has implemented
// the interface ServerWrapper and unwrap it.
// If still failing, return 1 instead.
func GetServerWeight(server Server) int {
	switch s := server.(type) {
	case WeightedServer:
		return s.Weight()

	case ServerWrapper:
		return GetServerWeight(s.Unwrap())

	default:
		return 1
	}
}

var (
	serverpool4   = sync.Pool{New: func() any { return make(Servers, 0, 4) }}
	serverpool8   = sync.Pool{New: func() any { return make(Servers, 0, 8) }}
	serverpool16  = sync.Pool{New: func() any { return make(Servers, 0, 16) }}
	serverpool32  = sync.Pool{New: func() any { return make(Servers, 0, 32) }}
	serverpool64  = sync.Pool{New: func() any { return make(Servers, 0, 64) }}
	serverpool128 = sync.Pool{New: func() any { return make(Servers, 0, 128) }}
)

// AcquireServers acquires a preallocated zero-length servers from the pool.
func AcquireServers(expectedMaxCap int) Servers {
	switch {
	case expectedMaxCap <= 4:
		return serverpool4.Get().(Servers)

	case expectedMaxCap <= 8:
		return serverpool8.Get().(Servers)

	case expectedMaxCap <= 16:
		return serverpool16.Get().(Servers)

	case expectedMaxCap <= 32:
		return serverpool32.Get().(Servers)

	case expectedMaxCap <= 64:
		return serverpool64.Get().(Servers)

	default:
		return serverpool128.Get().(Servers)
	}
}

// ReleaseServers releases the servers back into the pool.
func ReleaseServers(ss Servers) {
	for i, _len := 0, len(ss); i < _len; i++ {
		ss[i] = nil
	}

	ss = ss[:0]
	cap := cap(ss)
	switch {
	case cap < 8:
		serverpool4.Put(ss)

	case cap < 16:
		serverpool8.Put(ss)

	case cap < 32:
		serverpool16.Put(ss)

	case cap < 64:
		serverpool32.Put(ss)

	case cap < 128:
		serverpool64.Put(ss)

	default:
		serverpool128.Put(ss)
	}
}
