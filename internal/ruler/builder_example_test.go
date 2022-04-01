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

package ruler_test

import (
	"fmt"
	"net/http"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router"
	ruleroute "github.com/xgfone/go-apiserver/http/router/routes/ruler"
	"github.com/xgfone/go-apiserver/internal/ruler"
)

func ExampleBuild() {
	routeManger := ruleroute.NewRouteManager()

	// Set the builder of the matcher rule
	//
	// Notice: NewRouteManager has set it as the default builder of the matcher rule.
	//         Here is only show-how.
	routeManger.BuildMatcherRule = ruler.Build

	// Route 1
	routeManger.
		Rule("Method(`GET`) && Path(`/path1/{id}`)").              // Build the matcher
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // Set the handler
			c := reqresp.GetContext(w, r)
			fmt.Fprintf(w, "route1: %s", c.Data["id"])
		})

	// Route 2
	routeManger.
		Rule("Method(`GET`) && Path(`/path2/{id}`)").              // Build the matcher
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // Set the handler
			c := reqresp.GetContext(w, r)
			fmt.Fprintf(w, "route2: %s", c.Data["id"])
		})

	router := router.NewRouter(routeManger)
	http.ListenAndServe("127.0.0.1:80", router)

	// Open a terminal and run the program:
	// $ go run main.go
	//
	// Open another terminal and run the http client:
	// $ curl http://127.0.0.1/path1/123
	// route1: 123
	// $ curl http://127.0.0.1/path2/123
	// route2: 123
}
