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

package ruler

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/xgfone/go-apiserver/http/reqresp"
	matcher "github.com/xgfone/go-http-matcher"
)

var kvpool = sync.Pool{New: func() interface{} { return &kvswrapper{kvs: make([]kv, 0, 4)} }}

type kvswrapper struct{ kvs []kv }

func (w *kvswrapper) append(kv kv) { w.kvs = append(w.kvs, kv) }
func (w *kvswrapper) reset() *kvswrapper {
	clear(w.kvs)
	w.kvs = w.kvs[:0]
	return w
}

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

func (p urlPath) Match(r *http.Request) (ok bool) {
	if p.plen == 0 {
		if p.isPrefix {
			path := matcher.GetPath(r)
			if !strings.HasPrefix(path, p.rawPath) {
				return false
			}

			mlen := len(p.rawPath)
			return mlen == len(path) || path[mlen] == '/'
		}

		path := matcher.GetPath(r)
		if p.rawPath[len(p.rawPath)-1] != '/' {
			if _len := len(path); _len > 1 && path[_len-1] == '/' {
				path = path[:_len-1]
			}
		}
		return path == p.rawPath
	}

	args := kvpool.Get().(*kvswrapper)
	path := matcher.GetPath(r)

	var i int
	for ; i < p.plen && len(path) > 0; i++ {
		ap := p.paths[i]
		if len(ap.name) == 0 {
			if !strings.HasPrefix(path, ap.path) {
				kvpool.Put(args.reset())
				return false
			}

			path = path[len(ap.path):]
			continue
		}

		if index := strings.IndexByte(path, '/'); index == -1 {
			args.append(kv{key: ap.name, value: path})
			path = ""
		} else {
			args.append(kv{key: ap.name, value: path[:index]})
			path = path[index:]
		}
	}

	ok = i == p.plen
	if ok && !p.isPrefix {
		ok = len(path) == 0
	}

	if ok {
		if c := reqresp.GetContext(r.Context()); c != nil {
			for i, _len := 0, len(args.kvs); i < _len; i++ {
				c.Data[args.kvs[i].key] = args.kvs[i].value
			}
		}
	}

	kvpool.Put(args.reset())
	return
}

func buildPathMatcher(desc, path string, isPrefix bool) matcher.Matcher {
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
				panic(fmt.Errorf("no path parameter name at index between %d and %d", leftIndex, rightIndex))
			}

			p.paths = append(p.paths, argPath{path: path[:leftIndex]})
			p.paths = append(p.paths, argPath{name: name})
			path = path[rightIndex+1:]
		}
		p.plen = len(p.paths)
	}

	prio := matcher.PriorityPath
	if isPrefix {
		prio = matcher.PriorityPathPrefix
	}

	prefixlen := len(p.rawPath)
	if len(p.paths) > 0 && p.paths[0].name == "" {
		prefixlen = len(p.paths[0].path)
	}
	if prefixlen > 0 {
		prio *= prefixlen
	}

	return matcher.New(prio, desc, p.Match)
}

func newPathMatcher(path string) matcher.Matcher {
	desc := fmt.Sprintf("Path(`%s`)", path)
	return buildPathMatcher(desc, path, false)
}

func newPathPrefixMatcher(pathPrefix string) matcher.Matcher {
	desc := fmt.Sprintf("PathPrefix(`%s`)", pathPrefix)
	if pathPrefix == "/" {
		return matcher.New(matcher.PriorityPathPrefix, desc, matcher.AlwaysTrue)
	}
	return buildPathMatcher(desc, pathPrefix, true)
}
