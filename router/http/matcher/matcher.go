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

// Package matcher implements the some http request matchers.
package matcher

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/xgfone/go-apiserver/server"
	"github.com/xgfone/netaddr"
)

// MatchFunc is a function to match the request.
type MatchFunc func(old *http.Request) (new *http.Request, ok bool)

// Matcher is used to check whether the route rule matches the request.
type Matcher interface {
	fmt.Stringer

	// Priority is the priority of the matcher.
	//
	// The bigger the value, the higher the priority.
	Priority() int

	// Match is used to check whether the rule matches the request.
	Match(old *http.Request) (new *http.Request, ok bool)
}

// Matchers is a group of Matchers.
type Matchers []Matcher

func (ms Matchers) Len() int           { return len(ms) }
func (ms Matchers) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms Matchers) Less(i, j int) bool { return ms[j].Priority() < ms[i].Priority() }

// Priority returns the sum of the priorities of all the matchers.
func (ms Matchers) Priority() (priority int) {
	for i, _len := 0, len(ms); i < _len; i++ {
		priority += ms[i].Priority()
	}
	return
}

type matcher struct {
	prio  int
	desc  string
	match MatchFunc
}

func (m matcher) String() string                              { return m.desc }
func (m matcher) Priority() int                               { return m.prio }
func (m matcher) Match(r *http.Request) (*http.Request, bool) { return m.match(r) }

// New returns a new route matcher.
func New(priority int, description string, match MatchFunc) Matcher {
	return matcher{prio: priority, desc: description, match: match}
}

type notMatcher struct{ Matcher }

func (m notMatcher) String() string { return fmt.Sprintf("Not(%s)", m.Matcher.String()) }
func (m notMatcher) Match(r *http.Request) (*http.Request, bool) {
	r, ok := m.Matcher.Match(r)
	return r, !ok
}

// Not returns a NOT matcher based on the original matcher.
func Not(matcher Matcher) Matcher { return notMatcher{matcher} }

type andMatcher Matchers

func (m andMatcher) String() string {
	ms := Matchers(m)
	ss := make([]string, len(ms))
	for i, matcher := range ms {
		ss[i] = matcher.String()
	}
	return fmt.Sprintf("And(%s)", strings.Join(ss, ", "))
}

func (m andMatcher) Priority() int { return Matchers(m).Priority() }
func (m andMatcher) Match(r *http.Request) (*http.Request, bool) {
	var ok bool
	ms := Matchers(m)
	for i, _len := 0, len(ms); i < _len; i++ {
		if r, ok = ms[i].Match(r); !ok {
			return r, false
		}
	}
	return r, true
}

// And returns a new AND matcher.
func And(matchers ...Matcher) Matcher {
	switch len(matchers) {
	case 0:
		panic("AndMatcher: no the matcher")
	case 1:
		return matchers[0]
	}

	ms := make(Matchers, len(matchers))
	copy(ms, matchers)
	sort.Stable(ms)
	return andMatcher(ms)
}

type orMatcher Matchers

func (m orMatcher) String() string {
	ms := Matchers(m)
	ss := make([]string, len(ms))
	for i, matcher := range ms {
		ss[i] = matcher.String()
	}
	return fmt.Sprintf("Or(%s)", strings.Join(ss, ", "))
}

func (m orMatcher) Priority() int { return Matchers(m).Priority() }
func (m orMatcher) Match(r *http.Request) (*http.Request, bool) {
	var ok bool
	ms := Matchers(m)
	for i, _len := 0, len(ms); i < _len; i++ {
		if r, ok = ms[i].Match(r); ok {
			return r, true
		}
	}
	return r, false
}

// Or returns a new OR matcher.
func Or(matchers ...Matcher) Matcher {
	switch len(matchers) {
	case 0:
		panic("OrMatcher: no the matcher")
	case 1:
		return matchers[0]
	}

	ms := make(Matchers, len(matchers))
	copy(ms, matchers)
	sort.Stable(ms)
	return orMatcher(ms)
}

// Must returns the matcher when err is equal to nil. Or, panic with err.
func Must(matcher Matcher, err error) Matcher {
	if err != nil {
		panic(err)
	}
	return matcher
}

/// ----------------------------------------------------------------------- ///

const (
	prioHeaderRegexp = 1
	prioHeader       = 2
	prioQuery        = 2
	prioClientIP     = 3
	prioMethod       = 4
	prioPathRegexp   = 5
	prioPath         = 6
	prioHostRegexp   = 7
	prioHost         = 8
)

// Path returns a path matcher to match the request path accurately.
func Path(path string) (Matcher, error) {
	// TOOD: path parameters

	desc := fmt.Sprintf("Path(%s)", path)
	return New(prioPath, desc, func(r *http.Request) (*http.Request, bool) {
		return r, r.URL.Path == path
	}), nil
}

// PathPrefix returns a path prefix matcher to match the prefix
// of the request path.
func PathPrefix(pathPrefix string) (Matcher, error) {
	// TOOD: path parameters

	desc := fmt.Sprintf("PathPrefix(%s)", pathPrefix)
	return New(prioPath, desc, func(r *http.Request) (*http.Request, bool) {
		return r, strings.HasPrefix(r.URL.Path, pathPrefix)
	}), nil
}

// Method returns a method matcher to match the request method.
func Method(method string) (Matcher, error) {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodDelete,
		http.MethodPut, http.MethodPatch, http.MethodConnect, http.MethodTrace,
		http.MethodOptions:
	default:
		return nil, fmt.Errorf("unknown http method '%s'", method)
	}

	desc := fmt.Sprintf("Method(%s)", method)
	return New(prioPath, desc, func(r *http.Request) (*http.Request, bool) {
		return r, r.Method == method
	}), nil
}

// ClientIP returns a matcher to match the remote address of the request.
//
// Support that clientIP is an IP or CIDR, such as "1.2.3.4", "1.2.3.0/24".
func ClientIP(clientIP string) (Matcher, error) {
	var err error
	var ipnet netaddr.IPNetwork

	if strings.IndexByte(clientIP, '/') > -1 {
		ipnet, err = netaddr.NewIPNetwork(clientIP)
		if err != nil {
			return nil, err
		}

		clientIP = ""
	}

	desc := fmt.Sprintf("ClientIP(%s)", clientIP)
	return New(prioPath, desc, func(r *http.Request) (*http.Request, bool) {
		remoteIP, _ := server.SplitHostPort(r.RemoteAddr)
		if clientIP != "" {
			return r, remoteIP == clientIP
		}

		ip := net.ParseIP(remoteIP)
		if ip == nil {
			return r, false
		}

		version := 4
		if strings.IndexByte(remoteIP, '.') == -1 {
			version = 6
		}

		ipaddr, _ := netaddr.NewIPAddress(ip, version)
		return r, ipnet.HasIP(ipaddr)
	}), nil
}

// Query returns a qeury matcher to match the request query.
func Query(key, value string) (Matcher, error) {
	if key == "" {
		return nil, fmt.Errorf("the query key is empty")
	}

	desc := fmt.Sprintf("Query(%s=%s)", key, value)
	return New(prioPath, desc, func(r *http.Request) (*http.Request, bool) {
		var ok bool
		if value == "" {
			ok = r.URL.Query().Has(key)
		} else {
			ok = r.URL.Query().Get(key) == value
		}
		return r, ok
	}), nil
}

// Header returns a header matcher to match the request header.
func Header(key, value string) (Matcher, error) {
	if key == "" {
		return nil, fmt.Errorf("the header key is empty")
	}

	desc := fmt.Sprintf("Header(%s=%s)", key, value)
	return New(prioPath, desc, func(r *http.Request) (*http.Request, bool) {
		var ok bool
		if value == "" {
			ok = url.Values(r.Header).Has(key)
		} else {
			ok = url.Values(r.Header).Get(key) == value
		}
		return r, ok
	}), nil
}

// HeaderRegexp returns a header regexp matcher to match the request
// header by the regexp.
func HeaderRegexp(key, value string) (Matcher, error) {
	// TODO:)
	return nil, fmt.Errorf("not implemented header regexp")
}

// Host returns a host matcher to match the request host.
func Host(host string) (Matcher, error) {
	desc := fmt.Sprintf("Host(%s)", host)
	return New(prioPath, desc, func(r *http.Request) (*http.Request, bool) {
		var rhost string
		if r.TLS != nil && r.TLS.ServerName != "" {
			rhost = r.TLS.ServerName
		} else {
			rhost, _ = server.SplitHostPort(r.Host)
		}

		return r, rhost == host
	}), nil
}

// HostRegexp returns a host regexp matcher to match the request
// host by the regexp.
func HostRegexp(regexpHost string) (Matcher, error) {
	// TODO:)
	return nil, fmt.Errorf("not implemented host regexp")
}
