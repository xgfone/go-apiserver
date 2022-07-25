// Copyright 2022 xgfone
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

package vhost

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func httpHandler(statusCode int) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(statusCode)
	})
}

func TestManager(t *testing.T) {
	const (
		exactHost1 = "www.example1.com"
		exactHost2 = "www.example.com"
		suffixHost = "*.example.com"
	)

	m := NewManager()
	m.AddVHost(exactHost1, httpHandler(201))
	m.AddVHost(exactHost2, httpHandler(202))
	m.AddVHost(suffixHost, httpHandler(203))

	if vhosts := m.handler.Load().(vhostsWrapper).vhosts; len(vhosts) != 3 {
		t.Errorf("expect %d vhosts, but got %d: %v", 3, len(vhosts), vhosts)
	} else {
		for i, vhost := range vhosts {
			switch i {
			case 0:
				if vhost.VHost != exactHost1 {
					t.Errorf("expect vhost '%s', but got '%s'", exactHost1, vhost.VHost)
				}

			case 1:
				if vhost.VHost != exactHost2 {
					t.Errorf("expect vhost '%s', but got '%s'", exactHost2, vhost.VHost)
				}

			case 2:
				if vhost.VHost != ".example.com" {
					t.Errorf("expect vhost '%s', but got '%s'", ".example.com", vhost.VHost)
				}

			}
		}
	}

	if m.GetVHost(exactHost1) == nil {
		t.Errorf("not get the handler of the virtual host '%s'", exactHost1)
	}
	if m.GetVHost(exactHost2) == nil {
		t.Errorf("not get the handler of the virtual host '%s'", exactHost2)
	}
	if m.GetVHost(suffixHost) == nil {
		t.Errorf("not get the handler of the virtual host '%s'", suffixHost)
	}

	if vhosts := m.GetVHosts(); len(vhosts) != 3 {
		t.Errorf("expect %d vhosts, but got %d: %v", 3, len(vhosts), vhosts)
	} else {
		for vhost := range vhosts {
			switch vhost {
			case exactHost1, exactHost2, suffixHost:
			default:
				t.Errorf("unexpect vhost '%s'", vhost)
			}
		}
	}

	req1 := httptest.NewRequest(http.MethodGet, "http://www.example1.com", nil)
	rec1 := httptest.NewRecorder()
	m.ServeHTTP(rec1, req1)
	if rec1.Code != 201 {
		t.Errorf("expect status code %d, but got %d", 201, rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "http://www.example.com", nil)
	rec2 := httptest.NewRecorder()
	m.ServeHTTP(rec2, req2)
	if rec2.Code != 202 {
		t.Errorf("expect status code %d, but got %d", 202, rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "http://abc.example.com", nil)
	rec3 := httptest.NewRecorder()
	m.ServeHTTP(rec3, req3)
	if rec3.Code != 203 {
		t.Errorf("expect status code %d, but got %d", 203, rec3.Code)
	}

	req4 := httptest.NewRequest(http.MethodGet, "http://www.example3.com", nil)
	rec4 := httptest.NewRecorder()
	m.ServeHTTP(rec4, req4)
	if rec4.Code != 404 {
		t.Errorf("expect status code %d, but got %d", 404, rec4.Code)
	}

	rec4 = httptest.NewRecorder()
	m.SetDefaultVHost(httpHandler(204))
	m.ServeHTTP(rec4, req4)
	if rec4.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec4.Code)
	}

	rec4 = httptest.NewRecorder()
	m.SetDefaultVHost(nil)
	m.HandleHTTP = func(w http.ResponseWriter, _ *http.Request, _ http.Handler) {
		w.WriteHeader(205)
	}
	m.ServeHTTP(rec4, req4)
	if rec4.Code != 205 {
		t.Errorf("expect status code %d, but got %d", 205, rec4.Code)
	}
}

func ExampleManager() {
	httpHandler := func(body string) http.HandlerFunc {
		return func(rw http.ResponseWriter, _ *http.Request) {
			fmt.Fprintln(rw, body)
		}
	}

	vhosts := NewManager()
	vhosts.SetDefaultVHost(httpHandler("default"))

	// vhost1: www.example1.com
	vhost1 := http.NewServeMux()
	vhost1.HandleFunc("/path1", httpHandler("vhost1: /path1"))
	vhost1.HandleFunc("/path2", httpHandler("vhost1: /path2"))
	vhosts.AddVHost("www.example1.com", vhost1)

	// vhost2: www.example2.com
	vhost2 := http.NewServeMux()
	vhost2.HandleFunc("/path1", httpHandler("vhost2: /path1"))
	vhost2.HandleFunc("/path2", httpHandler("vhost2: /path2"))
	vhosts.AddVHost("www.example2.com", vhost2)

	// vhost3: *.example2.com
	vhost3 := http.NewServeMux()
	vhost3.HandleFunc("/path1", httpHandler("vhost3: /path1"))
	vhost3.HandleFunc("/path2", httpHandler("vhost3: /path2"))
	vhosts.AddVHost("*.example2.com", vhost3)

	// Start HTTP Server
	http.ListenAndServe("127.0.0.1:9300", vhosts)

	// Output Result:
	// $ curl http://127.0.0.1:9300/path1 -H 'Host: www.example1.com'
	// vhost1: /path1
	//
	// $ curl http://127.0.0.1:9300/path1 -H 'Host: www.example2.com'
	// vhost2: /path1
	//
	// $ curl http://127.0.0.1:9300/path1 -H 'Host: abc.example2.com'
	// vhost3: /path1
	//
	// $ curl http://127.0.0.1:9300/path1 -H 'Host: www.example3.com'
	// default
}
