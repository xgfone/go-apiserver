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

package matcher

import (
	"net/http"
	"testing"
)

func testMatcher(t *testing.T, req *http.Request, matcher Matcher, match bool) {
	if _, ok := matcher.Match(req); ok != match {
		t.Errorf("'%s': expect '%v', but got '%v'", matcher.String(), match, ok)
	}
}

func TestMatcher(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://www.example.com/path/to?v1=k1", nil)
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
	testMatcher(t, req, Must(Path("/path/to")), true)
	testMatcher(t, req, Must(Path("/")), false)
	// TODO: test path parameters

	// PathPrefix
	testMatcher(t, req, Must(PathPrefix("/path/")), true)
	testMatcher(t, req, Must(PathPrefix("/nopath")), false)
	// TODO: test path parameters

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
