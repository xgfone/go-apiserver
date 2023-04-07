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
)

func ExampleRouteBuilder_Group() {
	router := NewRouter()

	// ----- V1 Version -----
	v1 := router.Group("/v1")

	v1auth := v1.Group("/auth")
	v1auth.Path("/login").POSTFunc(func(w http.ResponseWriter, r *http.Request) { /* TODO */ })
	v1auth.Path("/logout").POSTFunc(func(w http.ResponseWriter, r *http.Request) { /* TODO */ })

	v1svc1 := v1.Group("/svc1")
	v1svc1.Path("/path").GETFunc(func(w http.ResponseWriter, r *http.Request) { /* TODO */ })

	v1svc2 := v1.Group("/svc2")
	v1svc2.Path("/path").GETFunc(func(w http.ResponseWriter, r *http.Request) { /* TODO */ })

	// ----- V2 Version -----
	v2 := router.Group("/v2")

	v2auth := v2.Group("/auth")
	v2auth.Path("/login").POSTFunc(func(w http.ResponseWriter, r *http.Request) { /* TODO */ })
	v2auth.Path("/logout").POSTFunc(func(w http.ResponseWriter, r *http.Request) { /* TODO */ })

	v2svc1 := v2.Group("/svc1")
	v2svc1.Path("/path").GETFunc(func(w http.ResponseWriter, r *http.Request) { /* TODO */ })

	v2svc2 := v2.Group("/svc2")
	v2svc2.Path("/path").GETFunc(func(w http.ResponseWriter, r *http.Request) { /* TODO */ })

	for _, route := range router.GetRoutes() {
		fmt.Println(route.Matcher.String())
	}

	// Output:
	// (Path(`/v2/auth/logout`) && Method(`POST`))
	// (Path(`/v1/auth/logout`) && Method(`POST`))
	// (Path(`/v2/auth/login`) && Method(`POST`))
	// (Path(`/v1/auth/login`) && Method(`POST`))
	// (Path(`/v2/svc2/path`) && Method(`GET`))
	// (Path(`/v2/svc1/path`) && Method(`GET`))
	// (Path(`/v1/svc2/path`) && Method(`GET`))
	// (Path(`/v1/svc1/path`) && Method(`GET`))
}