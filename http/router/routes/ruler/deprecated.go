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

import "github.com/xgfone/go-apiserver/http/router/ruler"

// Re-export the global variables and functions.
var (
	DefaultRouter = ruler.DefaultRouter

	NewRoute        = ruler.NewRoute
	NewRouter       = ruler.NewRouter
	NewRouteBuilder = ruler.NewRouteBuilder
)

// Re-export the types.
type (
	Route        = ruler.Route
	Routes       = ruler.Routes
	Router       = ruler.Router
	RouteBuilder = ruler.RouteBuilder
)
