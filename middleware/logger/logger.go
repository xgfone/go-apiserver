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

// LogRequest is used to log the request if set.
var LogRequest Logger

type (
	// Collector is used to collect the log key-value pairs.
	Collector func(kvs []interface{}) (newkvs []interface{}, clean func())

	// Logger is used to log the request.
	Logger func(ctx context.Context, req interface{}) Collector
)
