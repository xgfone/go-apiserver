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
	"github.com/xgfone/go-apiserver/internal/atomic"
	"github.com/xgfone/go-apiserver/log"
)

// ErrNoAvailableServers is used to represents no available servers.
var ErrNoAvailableServers = errors.New("no available servers")

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
	discoveryWrapper struct{ ServerDiscovery }
	forwarderWrapper struct{ Forwarder }
)

// LoadBalancer is used to forward the http request to one of the backend servers.
type LoadBalancer struct {
	name      string
	discovery atomic.Value
	forwarder atomic.Value
	timeout   int64

	handleError ErrorHandler

	slock   sync.RWMutex
	servers map[string]upstream.Server
	server  atomic.Value // For serversWrapper
}

// NewLoadBalancer returns a new LoadBalancer to forward the http request.
//
// If forwarder is nil, use Retry(RoundRobin()) by default.
//
// TODO: Add the retry when failed to forward the request.
func NewLoadBalancer(name string, forwarder Forwarder) *LoadBalancer {
	if forwarder == nil {
		forwarder = Retry(RoundRobin())
	}

	lb := &LoadBalancer{name: name, servers: make(map[string]upstream.Server, 8)}
	lb.server.Store(serversWrapper{})
	lb.forwarder.Store(forwarderWrapper{})
	lb.discovery.Store(discoveryWrapper{})
	lb.SetErrorHandler(nil)
	return lb
}

// Name reutrns the name of the upstream.
func (lb *LoadBalancer) Name() string { return lb.name }

// GetForwarder returns the forwarder.
func (lb *LoadBalancer) GetForwarder() Forwarder {
	return lb.forwarder.Load().(forwarderWrapper).Forwarder
}

// SwapForwarder swaps the old forwarder with the new.
func (lb *LoadBalancer) SwapForwarder(new Forwarder) (old Forwarder) {
	return lb.forwarder.Swap(forwarderWrapper{new}).(forwarderWrapper).Forwarder
}

// GetTimeout returns the maximum timeout.
func (lb *LoadBalancer) GetTimeout() time.Duration {
	return time.Duration(stdatomic.LoadInt64(&lb.timeout))
}

// SetTimeout sets the maximum timeout.
func (lb *LoadBalancer) SetTimeout(timeout time.Duration) {
	stdatomic.StoreInt64(&lb.timeout, int64(timeout))
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
		log.Info("forward the http request",
			log.F("upstream", lb.name),
			log.F("policy", lb.GetForwarder().Policy()),
			log.F("clientaddr", r.RemoteAddr),
			log.F("reqhost", r.Host),
			log.F("reqmethod", r.Method),
			log.F("reqpath", r.URL.Path))
		return

	case ErrNoAvailableServers:
		w.WriteHeader(503) // Service Unavailable

	default:
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			w.WriteHeader(504) // Gateway Timeout
		} else {
			w.WriteHeader(502) // Bad Gateway
		}
	}

	log.Error("fail to forward the http request",
		log.F("upstream", lb.name),
		log.F("policy", lb.GetForwarder().Policy()),
		log.F("clientaddr", r.RemoteAddr),
		log.F("reqhost", r.Host),
		log.F("reqmethod", r.Method),
		log.F("reqpath", r.URL.Path),
		log.E(err))
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
		return ErrNoAvailableServers
	}

	if timeout := lb.GetTimeout(); timeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		r = r.WithContext(ctx)
		defer cancel()
	}

	return lb.GetForwarder().Forward(w, r, servers)
}

// ServeHTTP implements the interface http.Handler.
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.handleError(w, r, lb.HandleHTTP(w, r))
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

func (lb *LoadBalancer) updateServers() {
	servers := make(upstream.Servers, 0, len(lb.servers))
	for _, server := range lb.servers {
		servers = append(servers, server)
	}
	sort.Stable(servers)
	lb.server.Store(serversWrapper{servers})
}

// ResetServers resets all the servers.
func (lb *LoadBalancer) ResetServers(servers ...upstream.Server) {
	lb.slock.Lock()
	defer lb.slock.Unlock()

	servermaps := make(map[string]upstream.Server, len(servers))
	for _len := len(servers) - 1; _len >= 0; _len-- {
		server := servers[_len]
		servermaps[server.ID()] = server
	}

	lb.servers = servermaps
	lb.updateServers()
}

// UpsertServers adds or updates the servers.
func (lb *LoadBalancer) UpsertServers(servers ...upstream.Server) {
	lb.slock.Lock()
	defer lb.slock.Unlock()

	for _len := len(servers) - 1; _len >= 0; _len-- {
		server := servers[_len]
		lb.servers[server.ID()] = server
	}
	lb.updateServers()
}

// RemoveServer removes and returns the server by the server id.
//
// If the server does not exist, do nothing and return nil.
func (lb *LoadBalancer) RemoveServer(id string) (server upstream.Server) {
	lb.slock.Lock()
	server, ok := lb.servers[id]
	if ok {
		delete(lb.servers, id)
		lb.updateServers()
	}
	lb.slock.Unlock()
	return
}

// GetServer returns the server by the server id.
func (lb *LoadBalancer) GetServer(id string) (server upstream.Server, ok bool) {
	lb.slock.RLock()
	server, ok = lb.servers[id]
	lb.slock.RUnlock()
	return
}

// GetServers returns all the servers.
func (lb *LoadBalancer) GetServers() upstream.Servers {
	servers := lb.server.Load().(serversWrapper).Servers
	newservers := make(upstream.Servers, len(servers))
	copy(newservers, servers)
	return newservers
}
