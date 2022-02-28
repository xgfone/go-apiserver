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

// Package middlewares is a collection of some middlewares.
package middlewares

import "github.com/xgfone/go-apiserver/http/middleware"

// DefaultMiddlewares is a set of the default middlewares.
var DefaultMiddlewares = middleware.Middlewares{Logger(1), Recover(10)}
