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
	"net/url"
	"sync"
)

const reqParamCtx reqParam = 0

type reqParam int8

// ReqCtx is used to represents the context information of the request.
type ReqCtx struct {
	// Req is the original http request.
	Req *http.Request

	Any   interface{}            // any single-value data
	Datas map[string]interface{} // a set of any key-value datas

	// Query are used to cache the parsed the form and the query.
	Query url.Values
}

// Queries parses and returns the query parameters.
func (r *ReqCtx) Queries() (query url.Values, err error) {
	if r.Query != nil {
		return r.Query, nil
	}

	if query, err = url.ParseQuery(r.Req.URL.RawQuery); err == nil {
		r.Query = query
	}

	return
}

// GetQueries is the same as Queries, but ingores the error.
func (r *ReqCtx) GetQueries() (query url.Values) {
	query, _ = r.Queries()
	return
}

// GetQuery parses the query parameters and return the value of the parameter
// by the key.
func (r *ReqCtx) GetQuery(key string) (value string) {
	return r.GetQueries().Get(key)
}

var reqCtxPool = sync.Pool{New: func() interface{} { return new(ReqCtx) }}

// NewReqCtx is used to creates the request context.
var NewReqCtx = AcquireReqCtx

// AcquireReqCtx acquires a request context from the pool.
func AcquireReqCtx(req *http.Request) *ReqCtx {
	reqCtx := reqCtxPool.Get().(*ReqCtx)
	reqCtx.Req = req
	return reqCtx
}

// ReleaseReqCtx releases the request context into the pool, which will reset
// all the fields of the request context to ZERO.
//
// If reqCtx is equal to nil, do nothing.
func ReleaseReqCtx(reqCtx *ReqCtx) {
	if reqCtx != nil {
		*reqCtx = ReqCtx{}
		reqCtxPool.Put(reqCtx)
	}
}

// SetReqCtx sets the request context into the request and returns a new one.
//
// If reqCtx is equal to nil, use NewReqCtx to create a new one.
func SetReqCtx(req *http.Request, reqCtx *ReqCtx) (newreq *http.Request) {
	if reqCtx == nil {
		reqCtx = NewReqCtx(req)
	}
	return req.WithContext(context.WithValue(req.Context(), reqParamCtx, reqCtx))
}

// GetReqCtx gets and returns the request context from the request.
//
// If the request context does not exist, reutrn nil.
func GetReqCtx(req *http.Request) *ReqCtx {
	if reqCtx, ok := req.Context().Value(reqParamCtx).(*ReqCtx); ok {
		return reqCtx
	}
	return nil
}

// GetReqDatas returns the all request parameters from the http request context.
func GetReqDatas(req *http.Request) (datas map[string]interface{}) {
	if reqCtx := GetReqCtx(req); reqCtx != nil {
		datas = reqCtx.Datas
	}
	return
}

// GetReqData returns the any request parameter by the key from the http request
// context.
//
// If the key does not exist, return nil.
func GetReqData(req *http.Request, key string) (value interface{}) {
	if reqCtx := GetReqCtx(req); reqCtx != nil && reqCtx.Datas != nil {
		value = reqCtx.Datas[key]
	}
	return
}

// SetReqData stores the any key-value request parameter into the http request
// context, and returns the new http request.
//
// If no request context is not set, use NewReqCtx to create a new one
// and store it into the new http request.
func SetReqData(req *http.Request, key string, value interface{}) (newreq *http.Request) {
	if key == "" {
		panic("the request parameter key is empty")
	}
	if value == nil {
		panic("the request parameter value is nil")
	}

	reqCtx := GetReqCtx(req)
	if reqCtx == nil {
		reqCtx = NewReqCtx(req)
		req = SetReqCtx(req, reqCtx)
	}

	if reqCtx.Datas == nil {
		reqCtx.Datas = make(map[string]interface{}, 8)
	}
	reqCtx.Datas[key] = value

	return req
}

// SetReqDatas is the same as SetReqData, but stores a set of key-value parameters.
func SetReqDatas(req *http.Request, datas map[string]interface{}) (newreq *http.Request) {
	if len(datas) == 0 {
		return req
	}

	reqCtx := GetReqCtx(req)
	if reqCtx == nil {
		reqCtx = NewReqCtx(req)
		req = SetReqCtx(req, reqCtx)
	}

	if reqCtx.Datas == nil {
		reqCtx.Datas = make(map[string]interface{}, 8+len(datas))
	}

	for key, value := range datas {
		reqCtx.Datas[key] = value
	}

	return req
}

// SetReqParam is the same as SetReqData, but assert the value to string.
func SetReqParam(req *http.Request, key, value string) (newreq *http.Request) {
	return SetReqData(req, key, value)
}

// SetReqParams is the same SetReqDatas, but assert the values to string.
func SetReqParams(req *http.Request, params map[string]string) (newreq *http.Request) {
	if len(params) == 0 {
		return req
	}

	reqCtx := GetReqCtx(req)
	if reqCtx == nil {
		reqCtx = NewReqCtx(req)
		req = SetReqCtx(req, reqCtx)
	}

	if reqCtx.Datas == nil {
		reqCtx.Datas = make(map[string]interface{}, 8+len(params))
	}

	for key, value := range params {
		reqCtx.Datas[key] = value
	}

	return req
}

// GetReqParam is the same as GetReqData, but assert the value to string.
func GetReqParam(req *http.Request, key string) (value string, ok bool) {
	value, ok = GetReqData(req, key).(string)
	return
}

// GetReqParams is the same as GetReqDatas, but assert the value to string.
//
// Suggest to use GetReqDatas instead of GetReqParams.
func GetReqParams(req *http.Request) (params map[string]string) {
	if reqCtx := GetReqCtx(req); reqCtx != nil {
		params = make(map[string]string, len(reqCtx.Datas))
		for key, value := range reqCtx.Datas {
			if v, ok := value.(string); ok {
				params[key] = v
			}
		}
	}
	return
}
