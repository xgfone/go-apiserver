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

// Package logger provides the global log configuration
// used by the middlewares.
package logger

import "context"

var (
	// LogRequest is used to log the request if set,
	//
	// If returning nil, do not log any extra information.
	//
	// For the default implementation, it returns nil.
	Start func(ctx context.Context, req interface{}) Collector = defaultStart

	// Enabled is used to decide whether to log the request,
	//
	// For the default implementation, it returns true.
	Enabled func(ctx context.Context, req interface{}) bool = defaultEnabled
)

// Collector is used to collect the extra log key-value pairs.
//
// If the returned clean function is nil, it indicates not to need to clean any.
type Collector func(kvs []interface{}) (newkvs []interface{}, clean func())

func defaultStart(ctx context.Context, req interface{}) Collector { return nil }
func defaultEnabled(ctx context.Context, req interface{}) bool    { return true }
