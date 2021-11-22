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

import (
	"context"
	"net/http"
)

const reqParamCtx reqParam = 0

type reqParam int8

// SetParam stores the key-value parameter into the http request context.
func SetParam(req *http.Request, key, value string) (newreq *http.Request) {
	ctx := req.Context()
	if params, ok := ctx.Value(reqParamCtx).(map[string]string); ok {
		params[key] = value
		return req
	}

	params := make(map[string]string, 8)
	params[key] = value
	return req.WithContext(context.WithValue(ctx, reqParamCtx, params))
}

// SetParams stores the key-value parameter into the http request context.
func SetParams(req *http.Request, vs map[string]string) (newreq *http.Request) {
	if len(vs) == 0 {
		return req
	}

	ctx := req.Context()
	if params, ok := ctx.Value(reqParamCtx).(map[string]string); ok {
		for key, value := range vs {
			params[key] = value
		}
		return req
	}

	params := make(map[string]string, 8+len(vs))
	for key, value := range vs {
		params[key] = value
	}
	return req.WithContext(context.WithValue(ctx, reqParamCtx, params))
}

// GetParam reads the parameter by the key from the http request context.
func GetParam(req *http.Request, key string) (value string, ok bool) {
	if vars, _ok := req.Context().Value(reqParamCtx).(map[string]string); _ok {
		value, ok = vars[key]
	}
	return
}

// GetParams reads all the parameters from the http request context.
//
// Notice: the returned map should be read-only.
func GetParams(req *http.Request) (params map[string]string) {
	params, _ = req.Context().Value(reqParamCtx).(map[string]string)
	return
}
