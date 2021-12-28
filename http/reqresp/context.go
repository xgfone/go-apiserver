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

package reqresp

import (
	"context"
	"net/http"
	"net/url"
	"sync"
)

// ContextAllocator is used to allocate or release the request context.
type ContextAllocator interface {
	Acquire(*http.Request) *Context
	Release(*Context)
}

type contextAllocator struct{ ctxPool sync.Pool }

func (a *contextAllocator) Acquire(req *http.Request) (c *Context) {
	c = a.ctxPool.Get().(*Context)
	c.Req = req
	return
}

func (a *contextAllocator) Release(c *Context) {
	if c != nil {
		// Clean the datas.
		if len(c.Datas) > 0 {
			for key := range c.Datas {
				delete(c.Datas, key)
			}
		}

		// Reset the any data.
		if reset, ok := c.Any.(interface{ Reset() }); ok {
			reset.Reset()
		}

		*c = Context{Datas: c.Datas, Any: c.Any}
		a.ctxPool.Put(c)
	}
}

// DefaultContextAllocator is the default request context allocator.
var DefaultContextAllocator = NewContextAllocator()

// NewContextAllocator returns a new ContextAllocator, which acquires a request
// context from the pool and releases the request context into the pool.
//
// Notice: if Context.Any has implemented the interface { Reset() },
// it will be called when releasing the request context.
func NewContextAllocator() ContextAllocator {
	var alloc contextAllocator
	alloc.ctxPool.New = func() interface{} { return new(Context) }
	return &alloc
}

type reqParam uint8

// SetContext sets the request context into the request and returns a new one.
//
// If c is equal to nil, do nothing and return the original http request.
func SetContext(req *http.Request, c *Context) (newreq *http.Request) {
	if c == nil {
		return req
	}
	return req.WithContext(context.WithValue(req.Context(), reqParam(255), c))
}

// GetContext gets and returns the request context from the request.
//
// If the request context does not exist, reutrn nil.
func GetContext(req *http.Request) *Context {
	if c, ok := req.Context().Value(reqParam(255)).(*Context); ok {
		return c
	}
	return nil
}

// GetOrNewContext is the same as GetContext, but create a new one
// if the request context does not exist.
func GetOrNewContext(req *http.Request) (c *Context, new bool) {
	if c = GetContext(req); c == nil {
		c = DefaultContextAllocator.Acquire(req)
		new = true
	}
	return
}

// GetReqDatas returns the all request parameters from the http request context.
func GetReqDatas(req *http.Request) (datas map[string]interface{}) {
	if c := GetContext(req); c != nil {
		datas = c.Datas
	}
	return
}

// GetReqData returns the any request parameter by the key from the http request
// context.
//
// If the key does not exist, return nil.
func GetReqData(req *http.Request, key string) (value interface{}) {
	if c := GetContext(req); c != nil && c.Datas != nil {
		value = c.Datas[key]
	}
	return
}

// SetReqData stores the any key-value request parameter into the http request
// context, and returns the new http request.
//
// If no request context is not set, use DefaultContextAllocator to create
// a new one and store it into the new http request.
func SetReqData(req *http.Request, key string, value interface{}) (newreq *http.Request) {
	if key == "" {
		panic("the request parameter key is empty")
	}
	if value == nil {
		panic("the request parameter value is nil")
	}

	c := GetContext(req)
	if c == nil {
		c = DefaultContextAllocator.Acquire(req)
		req = SetContext(req, c)
	}

	if c.Datas == nil {
		c.Datas = make(map[string]interface{}, 8)
	}
	c.Datas[key] = value

	return req
}

// SetReqDatas is the same as SetReqData, but stores a set of key-value parameters.
func SetReqDatas(req *http.Request, datas map[string]interface{}) (newreq *http.Request) {
	if len(datas) == 0 {
		return req
	}

	c := GetContext(req)
	if c == nil {
		c = DefaultContextAllocator.Acquire(req)
		req = SetContext(req, c)
	}

	if c.Datas == nil {
		c.Datas = make(map[string]interface{}, 8+len(datas))
	}

	for key, value := range datas {
		c.Datas[key] = value
	}

	return req
}

// Context is used to represents the context information of the request.
type Context struct {
	// Req and Resp are the http request and response.
	Resp ResponseWriter
	Req  *http.Request

	Any   interface{}            // any single-value data
	Datas map[string]interface{} // a set of any key-value datas

	// Query are used to cache the parsed request query.
	Query url.Values
}

// ParseQuery parses the query parameters, caches and returns the parsed query.
func (c *Context) ParseQuery() (query url.Values, err error) {
	if c.Query != nil {
		return c.Query, nil
	}

	if query, err = url.ParseQuery(c.Req.URL.RawQuery); err == nil {
		c.Query = query
	}

	return
}

// GetQueries is the same as Queries, but ingores the error.
func (c *Context) GetQueries() (query url.Values) {
	query, _ = c.ParseQuery()
	return
}

// GetQuery parses the query parameters and return the value of the parameter
// by the key.
func (c *Context) GetQuery(key string) (value string) {
	return c.GetQueries().Get(key)
}
