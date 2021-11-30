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

package http

import "net/http"

// DefaultHandlerManager is the default global http handler manager.
var DefaultHandlerManager = NewHandlerManager()

// AddHandler is equal to DefaultHandlerManager.AddHandler(name, handler).
func AddHandler(name string, handler http.Handler) (err error) {
	return DefaultHandlerManager.AddHandler(name, handler)
}

// DelHandler is equal to DefaultHandlerManager.DelHandler(name).
func DelHandler(name string) http.Handler {
	return DefaultHandlerManager.DelHandler(name)
}

// GetHandler is equal to DefaultHandlerManager.GetHandler(name).
func GetHandler(name string) http.Handler {
	return DefaultHandlerManager.GetHandler(name)
}

// GetHandlers is equal to DefaultHandlerManager.GetHandlers().
func GetHandlers() map[string]http.Handler {
	return DefaultHandlerManager.GetHandlers()
}

/// ----------------------------------------------------------------------- ///

// DefaultMiddlewareManager is the default global middleware manager.
var DefaultMiddlewareManager = NewMiddlewareManager()

// AddMiddleware is equal to DefaultMiddlewareManager.AddMiddleware(mw).
func AddMiddleware(mw Middleware) (err error) {
	return DefaultMiddlewareManager.AddMiddleware(mw)
}

// DelMiddleware is equal to DefaultMiddlewareManager.DelMiddleware(name).
func DelMiddleware(name string) Middleware {
	return DefaultMiddlewareManager.DelMiddleware(name)
}

// GetMiddleware is equal to DefaultMiddlewareManager.GetMiddleware(name).
func GetMiddleware(name string) Middleware {
	return DefaultMiddlewareManager.GetMiddleware(name)
}

// GetMiddlewares is equal to DefaultMiddlewareManager.GetMiddlewares().
func GetMiddlewares() Middlewares {
	return DefaultMiddlewareManager.GetMiddlewares()
}

/// ----------------------------------------------------------------------- ///

// DefaultMiddlewareBuilderManager is the default global middleware buidler manager.
var DefaultMiddlewareBuilderManager = NewMiddlewareBuilderManager()

// BuildMiddleware is equal to DefaultMiddlewareBuilderManager.Build(typ, name, conf).
func BuildMiddleware(typ, name string, conf map[string]interface{}) (Middleware, error) {
	return DefaultMiddlewareBuilderManager.Build(typ, name, conf)
}

// AddMiddlewareBuilder is equal to DefaultMiddlewareBuilderManager.AddBuilder(b).
func AddMiddlewareBuilder(b MiddlewareBuilder) (err error) {
	return DefaultMiddlewareBuilderManager.AddBuilder(b)
}

// DelMiddlewareBuilder is equal to DefaultMiddlewareBuilderManager.DelBuilder(name).
func DelMiddlewareBuilder(name string) MiddlewareBuilder {
	return DefaultMiddlewareBuilderManager.DelBuilder(name)
}

// GetMiddlewareBuilder is equal to DefaultMiddlewareBuilderManager.GetBuilder(name).
func GetMiddlewareBuilder(name string) MiddlewareBuilder {
	return DefaultMiddlewareBuilderManager.GetBuilder(name)
}

// GetMiddlewareBuilders is equal to DefaultMiddlewareBuilderManager.GetBuilders().
func GetMiddlewareBuilders() []MiddlewareBuilder {
	return DefaultMiddlewareBuilderManager.GetBuilders()
}
