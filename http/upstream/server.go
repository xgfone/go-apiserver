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

// Package upstream provides some common upstream functions.
package upstream

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	"github.com/xgfone/go-apiserver/nets"
)

// URL is the metadata information of the http endpoint.
type URL struct {
	Method  string            `json:"method" yaml:"method"` // Such as "GET"
	Scheme  string            `json:"scheme" yaml:"scheme"` // Such as "http" or "https"
	Domain  string            `json:"domain" yaml:"domain"` // Such as "www.example.com"
	IP      string            `json:"ip" yaml:"ip"`         // Such as "1.2.3.4"
	Port    uint16            `json:"port" yaml:"port"`     // Such as 80 or 443
	Path    string            `json:"path" yaml:"path"`     // Such as "/"
	Queries map[string]string `json:"queries" yaml:"queries"`
	Headers map[string]string `json:"headers" yaml:"headers"`
}

// Equal reports whether the url is equal to other.
func (u URL) Equal(other URL) bool { return reflect.DeepEqual(u, other) }

// IsZero reports whether the url is ZERO.
func (u URL) IsZero() bool {
	return u.Port == 0 &&
		len(u.IP) == 0 &&
		len(u.Path) == 0 &&
		len(u.Method) == 0 &&
		len(u.Scheme) == 0 &&
		len(u.Domain) == 0 &&
		len(u.Queries) == 0 &&
		len(u.Headers) == 0
}

// String returns the URL string.
func (u URL) String() string { url := u.URL(); return url.String() }

// URL returns the stdlib url.URL.
func (u URL) URL() url.URL {
	_url := url.URL{Scheme: u.Scheme, Path: u.Path}

	if u.IP != "" {
		if u.Port == 0 {
			_url.Host = u.IP
		} else {
			_url.Host = net.JoinHostPort(u.IP, fmt.Sprint(u.Port))
		}
	} else if u.Domain != "" {
		if u.Port == 0 {
			_url.Host = u.Domain
		} else {
			_url.Host = net.JoinHostPort(u.Domain, fmt.Sprint(u.Port))
		}
	} else {
		panic(fmt.Errorf("no url host: %+v", u))
	}

	if _len := len(u.Queries); _len > 0 {
		queries := make(url.Values, _len)
		for key, value := range u.Queries {
			queries[key] = []string{value}
		}
		_url.RawQuery = queries.Encode()
	}

	return _url
}

// Request converts the URL to a http request with the GET method.
func (u URL) Request(ctx context.Context) (*http.Request, error) {
	u.SetDefault()
	return http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
}

// SetDefault sets the url information to the default if not set.
//
//   Path: "/"
//   Method: "GET"
//   Scheme: "http"
//
func (u *URL) SetDefault() {
	if u.Path == "" {
		u.Path = "/"
	}
	if u.Method == "" {
		u.Method = http.MethodGet
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
}

// Server represents an upstream http server.
type Server interface {
	ID() string
	URL() URL
	State() nets.RuntimeState
	Check(ctx context.Context, healthURL URL) error
	HandleHTTP(http.ResponseWriter, *http.Request) error
}

// WeightedServer represents an upstream http server with the weight.
type WeightedServer interface {
	// Weight returns the weight of the server, which must be a positive integer.
	//
	// The bigger the value, the higher the weight.
	Weight() int

	Server
}

// Servers represents a group of the servers.
type Servers []Server

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
var DefaultServersPool = NewServerPool(16)

// ServersPool is used to allocate and recycle the server slice.
type ServersPool struct{ pool sync.Pool }

// NewServerPool returns a new servers pool.
func NewServerPool(defaultCap int) *ServersPool {
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
