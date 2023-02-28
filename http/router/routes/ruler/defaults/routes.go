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

// Package defaults is used to import the default routes based on the rule.
//
// When importing the package, it will adds some default routes
// into the default rule router.
package defaults

import (
	"github.com/xgfone/go-apiserver/http/router/routes/action"
	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
)

func init() { AddDefaultRoutes(ruler.DefaultRouter, "", nil) }

// AddDefaultRoutes adds some default routes into router.
//
// If action is nil, use action.DefaultRouter instead.
func AddDefaultRoutes(router *ruler.Router, pathPrefix string, action *action.Router) {
	router.AddDebugVarsRoute(pathPrefix)
	router.AddDebugProfileRoutes(pathPrefix)
	router.AddDebugActionRoute(pathPrefix, action)
	router.AddDebugRuleRoute(pathPrefix)
}
