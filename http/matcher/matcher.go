// Copyright 2021~2022 xgfone
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

// Package matcher defines the matcher of the http route,
// and provides some matcher implementations.
package matcher

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/nets"
)

// MatchFunc is a function to match the request.
type MatchFunc func(http.ResponseWriter, *http.Request) (ok bool)

// Matcher is used to check whether the route rule matches the request.
type Matcher interface {
	fmt.Stringer

	// Priority is the priority of the matcher.
	//
	// The bigger the value, the higher the priority.
	Priority() int

	// Match is used to check whether the rule matches the request.
	Match(http.ResponseWriter, *http.Request) (ok bool)
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

func (m matcher) String() string                                    { return m.desc }
func (m matcher) Priority() int                                     { return m.prio }
func (m matcher) Match(w http.ResponseWriter, r *http.Request) bool { return m.match(w, r) }

// New returns a new route matcher.
func New(priority int, description string, match MatchFunc) Matcher {
	return matcher{prio: priority, desc: description, match: match}
}

type notMatcher struct{ Matcher }

func (m notMatcher) String() string { return "!" + m.Matcher.String() }
func (m notMatcher) Match(w http.ResponseWriter, r *http.Request) bool {
	return !m.Matcher.Match(w, r)
}

// Not returns a NOT matcher based on the original matcher.
func Not(matcher Matcher) Matcher { return notMatcher{matcher} }

type andMatcher Matchers

// MarshalJSON implements the interface json.Marshaler.
func (m andMatcher) MarshalJSON() ([]byte, error) {
	return helper.EncodeJSONBytes(m.String())
}

func (m andMatcher) String() string {
	return formatMatchers(" && ", Matchers(m))
}

func (m andMatcher) Priority() int { return Matchers(m).Priority() }
func (m andMatcher) Match(w http.ResponseWriter, r *http.Request) bool {
	ms := Matchers(m)
	for i, _len := 0, len(ms); i < _len; i++ {
		if !ms[i].Match(w, r) {
			return false
		}
	}
	return true
}

// And returns a new AND matcher.
func And(matchers ...Matcher) Matcher {
	switch len(matchers) {
	case 0:
		panic("AndMatcher: no the matcher")
	case 1:
		return matchers[0]
	}

	ms := make(Matchers, 0, len(matchers))
	for _, m := range matchers {
		if andm, ok := m.(andMatcher); ok {
			ms = append(ms, Matchers(andm)...)
		} else {
			ms = append(ms, m)
		}
	}
	sort.Stable(ms)
	return andMatcher(ms)
}

type orMatcher Matchers

// MarshalJSON implements the interface json.Marshaler.
func (m orMatcher) MarshalJSON() ([]byte, error) {
	return helper.EncodeJSONBytes(m.String())
}

func (m orMatcher) String() string {
	return formatMatchers(" || ", Matchers(m))
}

func (m orMatcher) Priority() int { return Matchers(m).Priority() }
func (m orMatcher) Match(w http.ResponseWriter, r *http.Request) bool {
	ms := Matchers(m)
	for i, _len := 0, len(ms); i < _len; i++ {
		if ms[i].Match(w, r) {
			return true
		}
	}
	return false
}

// Or returns a new OR matcher.
func Or(matchers ...Matcher) Matcher {
	switch len(matchers) {
	case 0:
		panic("OrMatcher: no the matcher")
	case 1:
		return matchers[0]
	}

	ms := make(Matchers, 0, len(matchers))
	for _, m := range matchers {
		if orm, ok := m.(orMatcher); ok {
			ms = append(ms, Matchers(orm)...)
		} else {
			ms = append(ms, m)
		}
	}
	sort.Stable(ms)
	return orMatcher(ms)
}

func formatMatchers(sep string, matchers []Matcher) string {
	switch len(matchers) {
	case 0:
		return ""
	case 1:
		return matchers[0].String()
	}

	var b strings.Builder
	b.Grow(32)

	b.WriteByte('(')
	for i, matcher := range matchers {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(matcher.String())
	}
	b.WriteByte(')')

	return b.String()
}

// Must returns the matcher when err is equal to nil. Or, panic with err.
func Must(matcher Matcher, err error) Matcher {
	if err != nil {
		panic(err)
	}
	return matcher
}

/// ----------------------------------------------------------------------- ///

// GetPath is used to get the path from the request.
var GetPath = func(r *http.Request) (path string) { return r.URL.Path }

// GetHost is used to get the host name without the port from the request.
var GetHost = func(r *http.Request) (host string) {
	if r.TLS != nil && r.TLS.ServerName != "" {
		host = r.TLS.ServerName
	} else {
		host, _ = nets.SplitHostPort(r.Host)
	}
	return
}

// Pre-define some matcher priorities.
const (
	PriorityHeaderRegexp = 1
	PriorityHeader       = 2
	PriorityQuery        = 2
	PriorityClientIP     = 3
	PriorityMethod       = 4
	PriorityPathPrefix   = 5
	PriorityPath         = 6
	PriorityHostRegexp   = 7
	PriorityHost         = 8
)

var (
	// Path returns a path matcher to match the request path accurately.
	//
	// For the default implementation, it supports the path parameters,
	// such as "/{param1}/{param2}" or "/prefix/{param1}/path/{param2}/to".
	//
	// For the path arguments extracted from the request path, they will be
	// put into the Data field of reqresp.Context which is took out by the
	// function reqresp.GetContext if reqresp.Context exists.
	Path func(path string) (Matcher, error) = pathMatcher

	// PathPrefix returns a path prefix matcher to match the prefix
	// of the request path.
	//
	// For the default implementation, it supports the path parameters,
	// such as "/{param1}/{param2}" or "/prefix/{param1}/path/{param2}/to".
	// Furthermore, the prefix path "/prefix" matches the path "/prefix/",
	// but the prefix path "/prefix/" does not match the path "/prefix".
	//
	// For the path arguments extracted from the request path, they will be
	// put into the Data field of reqresp.Context which is took out by the
	// function reqresp.GetContext if reqresp.Context exists.
	PathPrefix func(pathPrefix string) (Matcher, error) = pathPrefixMatcher

	// Method returns a method matcher to match the request method.
	Method func(method string) (Matcher, error) = methodMatcher

	// ClientIP returns a matcher to match the remote address of the request.
	//
	// Support that clientIP is an IP or CIDR, such as "1.2.3.4", "1.2.3.0/24".
	ClientIP func(clientIP string) (Matcher, error) = clientIPMatcher

	// Query returns a qeury matcher to match the request query.
	//
	// If the value is empty, check whether the request contains the query "key".
	Query func(key, value string) (Matcher, error) = queryMatcher

	// Header returns a header matcher to match the request header.
	//
	// If the value is empty, check whether the request contains the header "key".
	Header func(key, value string) (Matcher, error) = headerMatcher

	// HeaderRegexp returns a header regexp matcher to match the request header
	// by the regexp.
	//
	// The default implementation uses the stdlib "regexp".
	HeaderRegexp func(key, regexpValue string) (Matcher, error) = headerRegexpMatcher

	// Host returns a host matcher to match the request host.
	Host func(host string) (Matcher, error) = hostMatcher

	// HostRegexp returns a host regexp matcher to match the request host
	// by the regexp.
	//
	// The default implementation uses the stdlib "regexp".
	HostRegexp func(regexpHost string) (Matcher, error) = hostRegexpMatcher
)

var kvpool = sync.Pool{New: func() interface{} { return make([]kv, 0, 4) }}

type kv struct {
	key   string
	value string
}

type argPath struct {
	name string
	path string
}

type urlPath struct {
	isPrefix bool
	rawPath  string
	paths    []argPath
	plen     int
}

func (p urlPath) Match(w http.ResponseWriter, r *http.Request) (ok bool) {
	if p.plen == 0 {
		if p.isPrefix {
			return strings.HasPrefix(GetPath(r), p.rawPath)
		}

		path := GetPath(r)
		if p.rawPath[len(p.rawPath)-1] != '/' {
			if _len := len(path); _len > 1 && path[_len-1] == '/' {
				path = path[:_len-1]
			}
		}
		return path == p.rawPath
	}

	args := kvpool.Get().([]kv)
	path := GetPath(r)

	var i int
	for ; i < p.plen && len(path) > 0; i++ {
		ap := p.paths[i]
		if len(ap.name) == 0 {
			if !strings.HasPrefix(path, ap.path) {
				kvpool.Put(args[:0])
				return false
			}

			path = path[len(ap.path):]
			continue
		}

		if index := strings.IndexByte(path, '/'); index == -1 {
			args = append(args, kv{key: ap.name, value: path})
			path = ""
		} else {
			args = append(args, kv{key: ap.name, value: path[:index]})
			path = path[index:]
		}
	}

	ok = i == p.plen
	if ok && !p.isPrefix {
		ok = len(path) == 0
	}

	if ok {
		if c := reqresp.GetContext(w, r); c != nil {
			for i, _len := 0, len(args); i < _len; i++ {
				c.Data[args[i].key] = args[i].value
			}
		}
	}

	kvpool.Put(args[:0])
	return
}

func newPathMatcher(desc, path string, isPrefix bool) (Matcher, error) {
	p := urlPath{isPrefix: isPrefix, rawPath: path}

	if strings.IndexByte(path, '{') > -1 && strings.IndexByte(path, '}') > -1 {
		p.paths = make([]argPath, 0, 4)
		for len(path) > 0 {
			leftIndex := strings.IndexByte(path, '{')
			if leftIndex == -1 {
				p.paths = append(p.paths, argPath{path: path})
				break
			}

			rightIndex := strings.IndexByte(path, '}')
			if rightIndex == -1 {
				p.paths = append(p.paths, argPath{path: path})
				break
			}

			name := path[leftIndex+1 : rightIndex]
			if name == "" {
				return nil, fmt.Errorf("no path parameter name at index between %d and %d",
					leftIndex, rightIndex)
			}

			p.paths = append(p.paths, argPath{path: path[:leftIndex]})
			p.paths = append(p.paths, argPath{name: name})
			path = path[rightIndex+1:]
		}
		p.plen = len(p.paths)
	}

	prio := PriorityPath
	if isPrefix {
		prio = PriorityPathPrefix
	}

	return New(prio, desc, p.Match), nil
}

func pathMatcher(path string) (Matcher, error) {
	if len(path) == 0 {
		return nil, errors.New("the url path is empty")
	} else if path[0] != '/' {
		return nil, fmt.Errorf("the url path does not start with '/'")
	}

	desc := fmt.Sprintf("Path(`%s`)", path)
	return newPathMatcher(desc, path, false)
}

func pathPrefixMatcher(pathPrefix string) (Matcher, error) {
	if len(pathPrefix) == 0 {
		return nil, errors.New("the url path prefix is empty")
	} else if pathPrefix[0] != '/' {
		return nil, fmt.Errorf("the url path prefix does not start with '/'")
	}

	desc := fmt.Sprintf("PathPrefix(`%s`)", pathPrefix)
	return newPathMatcher(desc, pathPrefix, true)
}

func methodMatcher(method string) (Matcher, error) {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodDelete,
		http.MethodPut, http.MethodPatch, http.MethodConnect, http.MethodTrace,
		http.MethodOptions:
	default:
		return nil, fmt.Errorf("unknown http method '%s'", method)
	}

	desc := fmt.Sprintf("Method(`%s`)", method)
	return New(PriorityMethod, desc, func(w http.ResponseWriter, r *http.Request) bool {
		return r.Method == method
	}), nil
}

func clientIPMatcher(clientIP string) (Matcher, error) {
	ipChecker, err := nets.NewIPChecker(clientIP)
	if err != nil {
		return nil, err
	}

	desc := fmt.Sprintf("ClientIP(`%s`)", clientIP)
	return New(PriorityClientIP, desc, func(_ http.ResponseWriter, r *http.Request) bool {
		remoteIP, _ := nets.SplitHostPort(r.RemoteAddr)
		return ipChecker.CheckIP(net.ParseIP(remoteIP))
	}), nil
}

func queryMatcher(key, value string) (Matcher, error) {
	if key == "" {
		return nil, fmt.Errorf("the query key is empty")
	}

	var desc string
	if value == "" {
		desc = fmt.Sprintf("Query(`%s`)", key)
	} else {
		desc = fmt.Sprintf("Query(`%s`, `%s`)", key, value)
	}

	return New(PriorityQuery, desc, func(w http.ResponseWriter, r *http.Request) bool {
		var queries url.Values
		if c := reqresp.GetContext(w, r); c != nil {
			queries = c.GetQueries()
		} else {
			queries = r.URL.Query()
		}

		var ok bool
		if value == "" {
			_, ok = queries[key]
		} else {
			ok = queries.Get(key) == value
		}
		return ok
	}), nil
}

func headerMatcher(key, value string) (Matcher, error) {
	if key == "" {
		return nil, fmt.Errorf("the header key is empty")
	}

	var desc string
	if value == "" {
		desc = fmt.Sprintf("Header(`%s`)", key)
	} else {
		desc = fmt.Sprintf("Header(`%s`, `%s`)", key, value)
	}

	return New(PriorityHeader, desc, func(_ http.ResponseWriter, r *http.Request) bool {
		var ok bool
		if value == "" {
			_, ok = r.Header[textproto.CanonicalMIMEHeaderKey(key)]
		} else {
			ok = r.Header.Get(key) == value
		}
		return ok
	}), nil
}

func headerRegexpMatcher(key, regexpValue string) (Matcher, error) {
	regexp, err := regexp.Compile(regexpValue)
	if err != nil {
		return nil, err
	}

	desc := fmt.Sprintf("HeaderRegexp(`%s`)", regexpValue)
	return New(PriorityHeaderRegexp, desc, func(w http.ResponseWriter, r *http.Request) bool {
		return regexp.MatchString(r.Header.Get(key))
	}), nil
}

func hostMatcher(host string) (Matcher, error) {
	desc := fmt.Sprintf("Host(`%s`)", host)
	return New(PriorityHost, desc, func(w http.ResponseWriter, r *http.Request) bool {
		return GetHost(r) == host
	}), nil
}

func hostRegexpMatcher(regexpHost string) (Matcher, error) {
	regexp, err := regexp.Compile(regexpHost)
	if err != nil {
		return nil, err
	}

	desc := fmt.Sprintf("HostRegexp(`%s`)", regexpHost)
	return New(PriorityHostRegexp, desc, func(w http.ResponseWriter, r *http.Request) bool {
		return regexp.MatchString(GetHost(r))
	}), nil
}
