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

package entrypoint

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/xgfone/go-apiserver/http/handler"
)

func TestHTTPEntryPoint(t *testing.T) {
	manager := NewManager()

	ep1, err := NewEntryPoint("http8001", "127.0.0.1:8001", handler.Handler200)
	if err != nil {
		t.Fatal(err)
	}

	ep2, err := NewEntryPoint("http8002", "127.0.0.1:8002", handler.Handler200)
	if err != nil {
		t.Fatal(err)
	}

	go ep1.Start()
	go ep2.Start()

	time.Sleep(time.Millisecond * 10)

	// Test HTTP Request
	resp, err := http.Get("http://127.0.0.1:8001")
	if err != nil {
		t.Error(err)
	}
	if resp != nil {
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Errorf("expect status code '%d', but got '%d'", 200, resp.StatusCode)
		}
	}

	resp, err = http.Get("http://127.0.0.1:8002")
	if err != nil {
		t.Error(err)
	}
	if resp != nil {
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Errorf("expect status code '%d', but got '%d'", 200, resp.StatusCode)
		}
	}

	// Test Manager

	manager.AddEntryPoint(ep1)
	manager.AddEntryPoint(ep2)

	ep := manager.GetEntryPoint(ep1.Name)
	if ep == nil || ep.Name != ep1.Name {
		t.Errorf("cannot get the entrypoint '%s'", ep1.Name)
	}

	eps := manager.GetEntryPoints()
	if len(eps) != 2 {
		t.Errorf("expect %d entrypoints, but got %d", 2, len(eps))
	} else {
		for _, ep := range eps {
			switch ep.Name {
			case eps[0].Name, eps[1].Name:
			default:
				t.Errorf("unexpected entrypoint '%s'", ep.Name)
			}
		}
	}

	if ep := manager.DelEntryPoint(ep1.Name); ep != nil {
		ep.Shutdown(context.Background())
	}

	if ep := manager.DelEntryPoint(ep2.Name); ep != nil {
		ep.Shutdown(context.Background())
	}

	eps = manager.GetEntryPoints()
	if len(eps) > 0 {
		t.Errorf("unexpect %d entrypoints", len(eps))
	}
}
