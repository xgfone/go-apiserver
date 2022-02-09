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
	"net/http"

	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
)

// DefaultManager is the default global entrypoint manager.
var DefaultManager = NewManager()

// AddEntryPoint is equal to DefaultManager.AddEntryPoint(ep).
func AddEntryPoint(ep *EntryPoint) error {
	return DefaultManager.AddEntryPoint(ep)
}

// DelEntryPoint is equal to DefaultManager.DelEntryPoint(name).
func DelEntryPoint(name string) *EntryPoint {
	return DefaultManager.DelEntryPoint(name)
}

// GetEntryPoint is equal to DefaultManager.GetEntryPoint(name).
func GetEntryPoint(name string) *EntryPoint {
	return DefaultManager.GetEntryPoint(name)
}

// GetEntryPoints is equal to DefaultManager.GetEntryPoints().
func GetEntryPoints() []*EntryPoint {
	return DefaultManager.GetEntryPoints()
}

// StartHTTPServer is a simple convenient function to start a http server.
func StartHTTPServer(name, addr string, handler http.Handler) {
	if handler == nil {
		panic("the http handler must not be nil")
	}

	ep, err := NewHTTPEntryPoint(name, addr, handler)
	if err != nil {
		log.Fatal().Str("name", name).Str("addr", addr).Err(err).
			Printf("fail to start the http server")
	}

	atexit.Register(ep.Stop)
	ep.OnShutdown(atexit.Execute)
	ep.Start()
}
