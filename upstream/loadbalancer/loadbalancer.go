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

// Package loadbalancer implements an upstream forwarder based on loadbalancer.
package loadbalancer

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/nets"
	"github.com/xgfone/go-apiserver/upstream"
	"github.com/xgfone/go-apiserver/upstream/balancer"
	"github.com/xgfone/go-apiserver/upstream/healthcheck"
)

var _ healthcheck.Updater = &LoadBalancer{}

type (
	balancerWrapper  struct{ balancer.Balancer }
	discoveryWrapper struct{ upstream.ServerDiscovery }
)

// LoadBalancer is used to forward the http request to one of the backend servers.
type LoadBalancer struct {
	name      string
	timeout   int64
	balancer  atomic.Value
	discovery atomic.Value

	svrManager *serversManager
	httpError  func(*LoadBalancer, *reqresp.Context, error)
}

// NewLoadBalancer returns a new LoadBalancer to forward the http request.
func NewLoadBalancer(name string, balancer balancer.Balancer) *LoadBalancer {
	if balancer == nil {
		panic("the balancer is nil")
	}

	lb := &LoadBalancer{name: name, svrManager: newServersManager()}
	lb.balancer.Store(balancerWrapper{balancer})
	lb.discovery.Store(discoveryWrapper{})
	lb.SetHTTPErrorHandler(nil)
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
	return time.Duration(atomic.LoadInt64(&lb.timeout))
}

// SetTimeout sets the maximum timeout.
func (lb *LoadBalancer) SetTimeout(timeout time.Duration) {
	atomic.StoreInt64(&lb.timeout, int64(timeout))
}

// GetServerDiscovery returns the server discovery.
//
// If not set the server discovery, return nil.
func (lb *LoadBalancer) GetServerDiscovery() (sd upstream.ServerDiscovery) {
	return lb.discovery.Load().(discoveryWrapper).ServerDiscovery
}

// SwapServerDiscovery sets the server discovery to discover the servers,
// and returns the old one.
//
// If sd is equal to nil, it will cancel the server discovery.
// Or, use the server discovery instead of the direct servers.
func (lb *LoadBalancer) SwapServerDiscovery(new upstream.ServerDiscovery) (old upstream.ServerDiscovery) {
	old = lb.discovery.Swap(discoveryWrapper{new}).(discoveryWrapper).ServerDiscovery
	// lb.ResetServers() // We need to clear all the servers??
	return
}

// SetHTTPErrorHandler sets the error handler to handle the error of forwarding
// the http request, so it may be used to log the request.
//
// If handler is equal to nil, reset it to the default.
func (lb *LoadBalancer) SetHTTPErrorHandler(handler func(*LoadBalancer, *reqresp.Context, error)) {
	if handler == nil {
		lb.httpError = handleHTTPError
	} else {
		lb.httpError = handler
	}
}

func handleHTTPError(lb *LoadBalancer, c *reqresp.Context, err error) {
	switch err {
	case nil:
		if log.Enabled(log.LevelDebug) {
			log.Debug("forward the http request",
				"requestid", upstream.GetRequestID(c.Context(), c.Request),
				"upstream", lb.name,
				"balancer", lb.GetBalancer().Policy(),
				"clientaddr", c.RemoteAddr,
				"reqhost", c.Host,
				"reqmethod", c.Method,
				"reqpath", c.URL.Path)
		}
		return

	case upstream.ErrNoAvailableServers:
		c.WriteHeader(503) // Service Unavailable

	default:
		if nets.IsTimeout(err) {
			c.WriteHeader(504) // Gateway Timeout
		} else {
			c.WriteHeader(502) // Bad Gateway
		}
	}

	log.Error("fail to forward the http request",
		"requestid", upstream.GetRequestID(c.Context(), c.Request),
		"upstream", lb.name,
		"balancer", lb.GetBalancer().Policy(),
		"clientaddr", c.RemoteAddr,
		"reqhost", c.Host,
		"reqmethod", c.Method,
		"reqpath", c.URL.Path,
		"err", err)
}

// ServeHTTP implements the interface http.Handler,
// which is the encapsulation of Serve for HTTP.
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := reqresp.GetContext(w, r)
	if c == nil {
		c = reqresp.DefaultContextAllocator.Acquire()
		c.ResponseWriter = reqresp.NewResponseWriter(w)
		c.Request = reqresp.SetContext(r, c)
		defer reqresp.DefaultContextAllocator.Release(c)
	}
	lb.httpError(lb, c, lb.Serve(c.Context(), c))
}

// Serve implement the interface upstream.Server#Serve,
// which will forward the request to one of the upstream servers.
func (lb *LoadBalancer) Serve(ctx context.Context, req interface{}) error {
	sd := lb.GetServerDiscovery()
	if sd == nil {
		sd = lb.svrManager
	}

	if sd.OnlineNum() <= 0 {
		return upstream.ErrNoAvailableServers
	}

	if timeout := lb.GetTimeout(); timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return lb.GetBalancer().Forward(ctx, req, sd)
}

// SetServerOnline implements the interface healthcheck.Updater#SetServerOnline
// to set the status of the server to ServerStatusOnline or ServerStatusOffline.
func (lb *LoadBalancer) SetServerOnline(serverID string, online bool) {
	if online {
		lb.SetServerStatus(serverID, upstream.ServerStatusOnline)
	} else {
		lb.SetServerStatus(serverID, upstream.ServerStatusOffline)
	}
}

// SetServerStatus sets the status of the server,
// which does nothing if the server does not exist.
//
// This is the inner server management of loadbalancer.
func (lb *LoadBalancer) SetServerStatus(serverID string, status upstream.ServerStatus) {
	lb.svrManager.SetServerStatus(serverID, status)
}

// SetServerStatuses sets the statuses of a set of servers,
// which does nothing if the server does not exist.
//
// This is the inner server management of loadbalancer.
func (lb *LoadBalancer) SetServerStatuses(statuses map[string]upstream.ServerStatus) {
	lb.svrManager.SetServerStatuses(statuses)
}

// ResetServers resets all the servers to servers with the status ServerStatusOnline.
//
// This is the inner server management of loadbalancer.
func (lb *LoadBalancer) ResetServers(servers ...upstream.Server) {
	lb.svrManager.ResetServers(servers...)
}

// UpsertServers adds or updates the servers.
//
// Notice: the statuses of all the given new servers are ServerStatusOnline.
//
// This is the inner server management of loadbalancer.
func (lb *LoadBalancer) UpsertServers(servers ...upstream.Server) {
	lb.svrManager.UpsertServers(servers...)
}

// UpsertServer adds or updates the given server with the status ServerStatusOnline.
//
// This is the inner server management of loadbalancer,
// which implements the interface healthcheck.Updater#UpsertServer.
func (lb *LoadBalancer) UpsertServer(server upstream.Server) {
	lb.svrManager.UpsertServers(server)
}

// RemoveServer removes the server by the server id,
// which does nothing if the server does not exist.
//
// This is the inner server management of loadbalancer,
// which implements the interface healthcheck.Updater#RemoveServer.
func (lb *LoadBalancer) RemoveServer(serverID string) {
	lb.svrManager.RemoveServer(serverID)
}

// GetServer returns the server by the server id.
//
// This is the inner server management of loadbalancer.
func (lb *LoadBalancer) GetServer(serverID string) (server upstream.Server, ok bool) {
	return lb.svrManager.GetServer(serverID)
}

// GetOnServers only returns all the online servers.
//
// This is the inner server management of loadbalancer.
func (lb *LoadBalancer) GetOnServers() upstream.Servers {
	return lb.svrManager.OnServers()
}

// GetOffServers only returns all the offline servers.
//
// This is the inner server management of loadbalancer.
func (lb *LoadBalancer) GetOffServers() upstream.Servers {
	return lb.svrManager.OffServers()
}

// GetAllServers returns all the servers, which are a set of the wrappers
// of the original servers and can be unwrapped by upstream.UnwrapServer.
//
// This is the inner server management of loadbalancer.
func (lb *LoadBalancer) GetAllServers() upstream.Servers {
	return lb.svrManager.AllServers()
}