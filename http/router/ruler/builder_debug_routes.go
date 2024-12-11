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
	"expvar"
	"net/http/pprof"
	rpprof "runtime/pprof"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

// DebugVars registers the vars route with the path "/debug/vars".
func (b RouteBuilder) DebugVars() RouteBuilder {
	return b.Path("/debug/vars").GET(expvar.Handler())
}

// DebugRuleRoutes registers the rule-routes route with the path "/debug/router/rule/routes".
//
// If router is nil, use DefaultRouter instead.
func (b RouteBuilder) DebugRuleRoutes(router *Router) RouteBuilder {
	return b.Path("/debug/router/rule/routes").GETContext(func(c *reqresp.Context) {
		var response struct {
			Routes []Route `json:"routes"`
		}
		if router == nil {
			response.Routes = DefaultRouter.Routes()
		} else {
			response.Routes = router.Routes()
		}
		c.JSON(200, response)
	})
}

// DebugProfiles registers the pprof routes with the path prefix "/debug/pprof/".
func (b RouteBuilder) DebugProfiles() RouteBuilder {
	router := b.Group("/debug/pprof")
	router.Path("/profile").GETFunc(pprof.Profile)
	router.Path("/cmdline").GETFunc(pprof.Cmdline)
	router.Path("/symbol").GETFunc(pprof.Symbol)
	router.Path("/trace").GETFunc(pprof.Trace)
	router.Path("/").GETFunc(pprof.Index)
	for _, p := range rpprof.Profiles() {
		router.Path(p.Name()).GET(pprof.Handler(p.Name()))
	}
	return b
}
