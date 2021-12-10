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

// Package loadbalancer implements an upstream forwarder based on loadbalancer.
package loadbalancer

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sort"
	"sync"
	stdatomic "sync/atomic"
	"time"

	"github.com/xgfone/go-apiserver/http/upstream"
	"github.com/xgfone/go-apiserver/http/upstream/balancer"
	"github.com/xgfone/go-apiserver/internal/atomic"
	"github.com/xgfone/go-apiserver/log"
)

// ErrorHandler is used to handle the error to respond to the original client.
//
// Notice: if the error is nil, it represents no error.
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

// ServerDiscovery is used to discover the servers.
type ServerDiscovery interface {
	Servers() upstream.Servers
	ID() string
}

type (
	serversWrapper   struct{ upstream.Servers }
	balancerWrapper  struct{ balancer.Balancer }
	discoveryWrapper struct{ ServerDiscovery }
)

type upserver struct {
	Server upstream.Server
	online int32
}

func newUpServer(server upstream.Server) *upserver {
	return &upserver{Server: server, online: 1}
}

func (s *upserver) IsOnline() bool {
	return stdatomic.LoadInt32(&s.online) == 1
}

func (s *upserver) SetOnline(online bool) (ok bool) {
	if online {
		ok = stdatomic.CompareAndSwapInt32(&s.online, 0, 1)
	} else {
		ok = stdatomic.CompareAndSwapInt32(&s.online, 1, 0)
	}
	return
}

// LoadBalancer is used to forward the http request to one of the backend servers.
type LoadBalancer struct {
	EnableLog bool

	name      string
	balancer  atomic.Value
	discovery atomic.Value
	timeout   int64

	handleError ErrorHandler

	slock   sync.RWMutex
	servers map[string]*upserver
	server  atomic.Value // For serversWrapper
}

// NewLoadBalancer returns a new LoadBalancer to forward the http request.
func NewLoadBalancer(name string, balancer balancer.Balancer) *LoadBalancer {
	if balancer == nil {
		panic("the balancer is nil")
	}

	lb := &LoadBalancer{name: name, servers: make(map[string]*upserver, 8)}
	lb.server.Store(serversWrapper{})
	lb.balancer.Store(balancerWrapper{balancer})
	lb.discovery.Store(discoveryWrapper{})
	lb.SetErrorHandler(nil)
	return lb
}

// Name reutrns the name of the loadbalander upstream.
func (lb *LoadBalancer) Name() string { return lb.name }

// Type reutrns the type of the loadbalander upstream.
func (lb *LoadBalancer) Type() string { return "loadbalancer" }

// GetBalancer returns the balancer.
func (lb *LoadBalancer) GetBalancer() balancer.Balancer {
	return lb.balancer.Load().(balancerWrapper).Balancer
}

// SwapBalancer swaps the old balancer with the new.
func (lb *LoadBalancer) SwapBalancer(new balancer.Balancer) (old balancer.Balancer) {
	return lb.balancer.Swap(balancerWrapper{new}).(balancerWrapper).Balancer
}

// GetTimeout returns the maximum timeout.
func (lb *LoadBalancer) GetTimeout() time.Duration {
	return time.Duration(stdatomic.LoadInt64(&lb.timeout))
}

// SetTimeout sets the maximum timeout.
func (lb *LoadBalancer) SetTimeout(timeout time.Duration) {
	stdatomic.StoreInt64(&lb.timeout, int64(timeout))
}

// GetServerDiscovery returns the server discovery.
//
// If not set the server discovery, return nil.
func (lb *LoadBalancer) GetServerDiscovery() (sd ServerDiscovery) {
	return lb.discovery.Load().(discoveryWrapper).ServerDiscovery
}

// SwapServerDiscovery sets the server discovery to discover the servers,
// and returns the old one.
//
// If sd is equal to nil, it will cancel the server discovery.
// Or use the server discovery instead of the direct servers.
func (lb *LoadBalancer) SwapServerDiscovery(new ServerDiscovery) (old ServerDiscovery) {
	old = lb.discovery.Swap(discoveryWrapper{new}).(discoveryWrapper).ServerDiscovery
	lb.ResetServers()
	return
}

// SetErrorHandler sets the error handler to respond to the original client.
//
// If handleError is equal to nil, reset it to the default.
func (lb *LoadBalancer) SetErrorHandler(handleError ErrorHandler) {
	if handleError == nil {
		lb.handleError = lb.errorHandler
	} else {
		lb.handleError = handleError
	}
}

func (lb *LoadBalancer) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	switch err {
	case nil:
		if lb.EnableLog {
			log.Debug("forward the http request",
				log.F("upstream", lb.name),
				log.F("balancer", lb.GetBalancer().Policy()),
				log.F("clientaddr", r.RemoteAddr),
				log.F("reqhost", r.Host),
				log.F("reqmethod", r.Method),
				log.F("reqpath", r.URL.Path))
		}
		return

	case upstream.ErrNoAvailableServers:
		w.WriteHeader(503) // Service Unavailable

	default:
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			w.WriteHeader(504) // Gateway Timeout
		} else {
			w.WriteHeader(502) // Bad Gateway
		}
	}

	if lb.EnableLog {
		log.Error("fail to forward the http request",
			log.F("upstream", lb.name),
			log.F("balancer", lb.GetBalancer().Policy()),
			log.F("clientaddr", r.RemoteAddr),
			log.F("reqhost", r.Host),
			log.F("reqmethod", r.Method),
			log.F("reqpath", r.URL.Path),
			log.E(err))
	}
}

// HandleHTTP implements the interface Server.
func (lb *LoadBalancer) HandleHTTP(w http.ResponseWriter, r *http.Request) error {
	var servers upstream.Servers
	if sd := lb.GetServerDiscovery(); sd != nil {
		servers = sd.Servers()
	} else {
		servers = lb.server.Load().(serversWrapper).Servers
	}

	if len(servers) == 0 {
		return upstream.ErrNoAvailableServers
	}

	if timeout := lb.GetTimeout(); timeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		r = r.WithContext(ctx)
		defer cancel()
	}

	return lb.GetBalancer().Forward(w, r, servers)
}

// ServeHTTP implements the interface http.Handler.
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.handleError(w, r, lb.HandleHTTP(w, r))
}

// SetServerOnline sets the online status of the server by its id.
//
// If the server does not exist, do nothing and return false.
func (lb *LoadBalancer) SetServerOnline(serverID string, online bool) (ok bool) {
	lb.slock.RLock()
	upserver, ok := lb.servers[serverID]
	if ok {
		if upserver.SetOnline(online) {
			lb.updateServers()
		}
	}
	lb.slock.RUnlock()

	return
}

// SetServerOnlines sets the online statuses of a set of the servers.
//
// If the server does not exist, do nothing.
func (lb *LoadBalancer) SetServerOnlines(onlines map[string]bool) {
	lb.slock.RLock()
	var changed bool
	for serverID, online := range onlines {
		if upserver, ok := lb.servers[serverID]; ok {
			if upserver.SetOnline(online) && !changed {
				changed = true
			}
		}
	}
	if changed {
		lb.updateServers()
	}
	lb.slock.RUnlock()
	return
}

// ServerIsOnline reports whether the server is online.
//
// If the server does not exist, return (false, false).
func (lb *LoadBalancer) ServerIsOnline(serverID string) (online, ok bool) {
	lb.slock.RLock()
	upserver, ok := lb.servers[serverID]
	lb.slock.RUnlock()

	if ok {
		online = upserver.IsOnline()
	}

	return
}

func (lb *LoadBalancer) updateServers() {
	servers := upstream.DefaultServersPool.Acquire()
	for _, server := range lb.servers {
		if server.IsOnline() {
			servers = append(servers, server.Server)
		}
	}
	sort.Stable(servers)

	old := lb.server.Swap(serversWrapper{servers})
	upstream.DefaultServersPool.Release(old.(serversWrapper).Servers)
}

// ResetServers resets all the servers.
//
// Notice: the online statuses of all the given servers are online.
func (lb *LoadBalancer) ResetServers(servers ...upstream.Server) {
	lb.slock.Lock()
	defer lb.slock.Unlock()

	servermaps := make(map[string]*upserver, len(servers))
	for _len := len(servers) - 1; _len >= 0; _len-- {
		server := servers[_len]
		servermaps[server.ID()] = newUpServer(server)
	}

	lb.servers = servermaps
	lb.updateServers()
}

// UpsertServers adds or updates the servers.
//
// Notice: the online statuses of all the given new servers are online.
func (lb *LoadBalancer) UpsertServers(servers ...upstream.Server) {
	lb.slock.Lock()
	defer lb.slock.Unlock()

	for _len := len(servers) - 1; _len >= 0; _len-- {
		server := servers[_len]
		lb.servers[server.ID()] = newUpServer(server)
	}
	lb.updateServers()
}

// RemoveServer removes and returns the server by the server id.
//
// If the server does not exist, do nothing and return nil.
func (lb *LoadBalancer) RemoveServer(id string) (server upstream.Server) {
	lb.slock.Lock()
	if upserver, ok := lb.servers[id]; ok {
		server = upserver.Server
		delete(lb.servers, id)
		lb.updateServers()
	}
	lb.slock.Unlock()
	return
}

// GetServer returns the server by the server id.
func (lb *LoadBalancer) GetServer(id string) (server upstream.Server, online, ok bool) {
	lb.slock.RLock()
	upserver, ok := lb.servers[id]
	lb.slock.RUnlock()

	if ok {
		server = upserver.Server
		online = upserver.IsOnline()
	}

	return
}

// GetServers returns all the servers with the online status.
func (lb *LoadBalancer) GetServers() (onlines map[upstream.Server]bool) {
	lb.slock.RLock()
	onlines = make(map[upstream.Server]bool, len(lb.servers))
	for _, upserver := range lb.servers {
		onlines[upserver.Server] = upserver.IsOnline()
	}
	lb.slock.RUnlock()
	return
}
