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

package matcher

import (
	"net/http"
	"net/url"
	"sort"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

func testMatcher(t *testing.T, req *http.Request, matcher Matcher, match bool) {
	if _, ok := matcher.Match(req); ok != match {
		t.Errorf("'%s': expect '%v', but got '%v'", matcher.String(), match, ok)
	}
}

func TestMatcher(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://www.example.com/path/to/?v1=k1", nil)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "1.2.3.4"

	// ClientIP
	testMatcher(t, req, Must(ClientIP("1.2.3.4")), true)
	testMatcher(t, req, Must(ClientIP("1.2.3.0/24")), true)
	testMatcher(t, req, Must(ClientIP("5.6.7.8")), false)

	// Method
	testMatcher(t, req, Must(Method("GET")), true)
	testMatcher(t, req, Must(Method("POST")), false)

	// Query
	testMatcher(t, req, Must(Query("v1", "k1")), true)
	testMatcher(t, req, Must(Query("v2", "k2")), false)

	// Host
	testMatcher(t, req, Must(Host("www.example.com")), true)
	testMatcher(t, req, Must(Host("www.example.cn")), false)

	// HostRegexp
	// TODO:)

	// Path
	testMatcher(t, req, Must(Path("/path/to/")), true)
	testMatcher(t, req, Must(Path("/path/to")), true)
	testMatcher(t, req, Must(Path("/")), false)

	// PathPrefix
	testMatcher(t, req, Must(PathPrefix("/path/")), true)
	testMatcher(t, req, Must(PathPrefix("/nopath")), false)

	// Header
	testMatcher(t, req, Must(Header("Content-Type", "application/json")), true)
	testMatcher(t, req, Must(Header("Content-Type", "application/xml")), false)

	// HeaderRegexp
	// TODO:)

	// Not
	testMatcher(t, req, Not(Must(ClientIP("1.2.3.4"))), false)

	// And
	testMatcher(t, req, And(Must(Method("GET")), Must(Path("/path/to"))), true)
	testMatcher(t, req, And(Must(Method("GET")), Must(Path("/path"))), false)

	// Or
	testMatcher(t, req, Or(Must(Method("GET")), Must(Path("/path/to"))), true)
	testMatcher(t, req, Or(Must(Method("GET")), Must(Path("/path"))), true)
	testMatcher(t, req, Or(Must(Method("POST")), Must(Path("/path"))), false)
}

func TestMatcherPriority(t *testing.T) {
	matchers := Matchers{
		New(1, "matcher1", nil),
		New(3, "matcher3", nil),
		New(2, "matcher2", nil),
		New(2, "matcher4", nil),
	}
	sort.Stable(matchers)

	for i, m := range matchers {
		if i == 0 && m.String() != "matcher3" {
			t.Errorf("%d: expect matcher '%s', but got '%s'", i, "matcher3", m.String())
		}
		if i == 1 && m.String() != "matcher2" {
			t.Errorf("%d: expect matcher '%s', but got '%s'", i, "matcher2", m.String())
		}
		if i == 2 && m.String() != "matcher4" {
			t.Errorf("%d: expect matcher '%s', but got '%s'", i, "matcher4", m.String())
		}
		if i == 3 && m.String() != "matcher1" {
			t.Errorf("%d: expect matcher '%s', but got '%s'", i, "matcher1", m.String())
		}
	}
}

func TestPathMatcherParameter(t *testing.T) {
	matchers := []Matcher{
		Must(Path("/prefix/{id}")),
		Must(Path("/prefix/{id}/")),
		Must(Path("/prefix/{id}/path")),
		Must(Path("/prefix/{id}/to/{name}")),
	}

	paths := []struct {
		Path string
		Args map[string]string
	}{
		{Path: "/prefix/123", Args: map[string]string{"id": "123"}},
		{Path: "/prefix/123/", Args: map[string]string{"id": "123"}},
		{Path: "/prefix/123/path", Args: map[string]string{"id": "123"}},
		{Path: "/prefix/123/to/abc", Args: map[string]string{"id": "123", "name": "abc"}},
	}

	req := &http.Request{URL: &url.URL{}}
	for i, m := range matchers {
		for j, p := range paths {
			if i == j {
				req.URL.Path = p.Path
				nreq, ok := m.Match(req)

				if !ok {
					t.Errorf("%s does not match the path '%s'", m.String(), p.Path)
					continue
				}

				if datas := reqresp.GetReqDatas(nreq); len(p.Args) != len(datas) {
					t.Errorf("expect %d arguments, but got %d: %v", len(p.Args), len(datas), datas)
					t.Error(reqresp.GetContext(nreq))
				} else {
					for key, value := range p.Args {
						if v := datas[key]; v != value {
							t.Errorf("argument '%s': expect value '%s', but got '%s'", key, value, v)
						}
					}
				}
			}
		}
	}
}

func TestPathPrefixMatcherParameter(t *testing.T) {
	matcher := Must(PathPrefix("/prefix/{id}"))

	paths := []struct {
		Match bool
		Path  string
		Args  map[string]string
	}{
		{Match: true, Path: "/prefix/123", Args: map[string]string{"id": "123"}},
		{Match: true, Path: "/prefix/123/", Args: map[string]string{"id": "123"}},
		{Match: true, Path: "/prefix/123/path", Args: map[string]string{"id": "123"}},
		{Match: false, Path: "/notmatch/123"},
	}

	req := &http.Request{URL: &url.URL{}}
	for _, p := range paths {
		req.URL.Path = p.Path
		nreq, ok := matcher.Match(req)
		if p.Match {
			if !ok {
				t.Errorf("%s does not match the path '%s'", matcher.String(), p.Path)
				continue
			}

			if datas := reqresp.GetReqDatas(nreq); len(p.Args) != len(datas) {
				t.Errorf("expect %d arguments, but got %d: %v", len(p.Args), len(datas), datas)
				t.Error(reqresp.GetContext(nreq))
			} else {
				for key, value := range p.Args {
					if v := datas[key]; v != value {
						t.Errorf("argument '%s': expect value '%s', but got '%s'", key, value, v)
					}
				}
			}
		} else {
			if ok {
				t.Errorf("%s does not expect to match the path '%s'", matcher.String(), p.Path)
			}
		}
	}

	matcher = Must(PathPrefix("/prefix/{id}/"))
	req.URL.Path = "/prefix/123"
	_, ok := matcher.Match(req)
	if ok {
		t.Errorf("unexpect the matcher '%s' to match the path '%s'", matcher.String(), req.URL.Path)
	}
}
