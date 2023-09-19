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
	"net/http"
	"net/url"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

func TestNewPathMatcher(t *testing.T) {
	req := &http.Request{URL: &url.URL{Path: "/"}}
	m := newPathMatcher("/")
	if !m.Match(req) {
		t.Errorf("expect match, but got not")
	}

	req.URL.Path = "/path/"
	if m.Match(req) {
		t.Errorf("unexpect match, but got true")
	}

	m = newPathMatcher("/path")
	if !m.Match(req) {
		t.Errorf("expect match, but got not")
	}

	c := reqresp.AcquireContext()
	req = req.WithContext(reqresp.SetContext(req.Context(), c))
	req.URL.Path = "/prefix/admin/123456/info"
	m = newPathMatcher("/prefix/{group}/{userid}/info")
	if !m.Match(req) {
		t.Errorf("expect match, but got not")
	} else if len(c.Data) != 2 {
		t.Errorf("expect %d arguments, but got %d", 2, len(c.Data))
	} else if group, _ := c.Data["group"].(string); group != "admin" {
		t.Errorf("expect group argument value '%s', but got '%s'", "admin", group)
	} else if userid, _ := c.Data["userid"].(string); userid != "123456" {
		t.Errorf("expect userid argument value '%s', but got '%s'", "123456", userid)
	}
}

func TestNewPathPrefixMatcher(t *testing.T) {
	req := &http.Request{URL: &url.URL{Path: "/"}}
	m := newPathPrefixMatcher("/")
	if !m.Match(req) {
		t.Errorf("expect match, but got not")
	}

	req.URL.Path = "/path/"
	if !m.Match(req) {
		t.Errorf("expect match, but got not")
	}

	m = newPathPrefixMatcher("/path")
	if !m.Match(req) {
		t.Errorf("expect match, but got not")
	}

	req.URL.Path = "/path/to"
	if !m.Match(req) {
		t.Errorf("unexpect match, but got true")
	}

	req.URL.Path = "/pathto"
	if m.Match(req) {
		t.Errorf("unexpect match, but got true")
	}

	c := reqresp.AcquireContext()
	req = req.WithContext(reqresp.SetContext(req.Context(), c))
	req.URL.Path = "/prefix/admin/123456/info"
	m = newPathPrefixMatcher("/prefix/{group}")
	if !m.Match(req) {
		t.Errorf("expect match, but got not")
	} else if len(c.Data) != 1 {
		t.Errorf("expect %d arguments, but got %d", 1, len(c.Data))
	} else if group, _ := c.Data["group"].(string); group != "admin" {
		t.Errorf("expect group argument value '%s', but got '%s'", "admin", group)
	}
}
