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

package middlewares_test

import (
	"net/http"

	"github.com/xgfone/go-apiserver/http/middlewares"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router"
	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
)

func ExampleCORS() {
	// Use the middleware CORS.
	router.DefaultRouter.Middlewares.Use(middlewares.CORS(123, nil))

	// Use the rule router to manage the routes, which has been done by default.
	router.DefaultRouter.RouteManager = ruler.DefaultRouter

	// Add the routes
	ruler.DefaultRouter.Path("/path/to/1").GET(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO:
	}))

	ruler.DefaultRouter.Path("/path/to/2").GETFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO:
	})

	ruler.DefaultRouter.Path("/path/to/3").GETContext(func(ctx *reqresp.Context) {
		// TODO:
	})

	ruler.DefaultRouter.Path("/path/to/4").GETContextWithError(func(ctx *reqresp.Context) error {
		// TODO:
		return nil
	})
}
