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

package reqresp

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/xgfone/go-apiserver/http/binder"
	"github.com/xgfone/go-apiserver/http/header"
)

type builder struct{ buf []byte }

func (b *builder) Reset() { b.buf = b.buf[:0] }

func (b *builder) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(b.buf)
	return int64(n), err
}

func (b *builder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *builder) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

var bpool = sync.Pool{New: func() interface{} {
	return &builder{make([]byte, 0, 1024)}
}}

func getBuilder() *builder  { return bpool.Get().(*builder) }
func putBuilder(b *builder) { b.Reset(); bpool.Put(b) }

/// ----------------------------------------------------------------------- ///

// ContextAllocator is used to allocate or release the request context.
type ContextAllocator interface {
	Acquire() *Context
	Release(*Context)
}

type contextAllocator struct{ ctxPool sync.Pool }

func (a *contextAllocator) Acquire() (c *Context) {
	return a.ctxPool.Get().(*Context)
}

func (a *contextAllocator) Release(c *Context) {
	if c != nil {
		c.Reset()
		a.ctxPool.Put(c)
	}
}

// DefaultContextAllocator is the default request context allocator.
var DefaultContextAllocator = NewContextAllocator(8)

// NewContextAllocator returns a new ContextAllocator, which acquires a request
// context from the pool and releases the request context into the pool.
//
// Notice: if Context.Any has implemented the interface { Reset() },
// it will be called when releasing the request context.
func NewContextAllocator(dataCap int) ContextAllocator {
	var alloc contextAllocator
	alloc.ctxPool.New = func() interface{} {
		return &Context{Datas: make(map[string]interface{}, dataCap)}
	}
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
		c = DefaultContextAllocator.Acquire()
		c.Request = req
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
		c = DefaultContextAllocator.Acquire()
		req = SetContext(req, c)
		c.Request = req
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
		c = DefaultContextAllocator.Acquire()
		req = SetContext(req, c)
		c.Request = req
	}

	if c.Datas == nil {
		c.Datas = make(map[string]interface{}, 8+len(datas))
	}

	for key, value := range datas {
		c.Datas[key] = value
	}

	return req
}

var _ ResponseWriter = &Context{}

// Context is used to represents the context information of the request.
type Context struct {
	ResponseWriter
	*http.Request

	Err    error                  // used to store the error
	Any    interface{}            // any single-value data
	Datas  map[string]interface{} // a set of any key-value datas
	Binder binder.Binder

	// Query and Cookies are used to cache the parsed request query and cookies.
	Cookies []*http.Cookie
	Query   url.Values
}

// Reset resets the context itself.
func (c *Context) Reset() {
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

	*c = Context{Any: c.Any, Datas: c.Datas, Binder: c.Binder}
}

// Bind extracts the data information from the request and assigns it to v.
func (c *Context) Bind(v interface{}) error {
	return c.Binder.Bind(v, c.Request)
}

// Header implements the interface ResponseWriter.
func (c *Context) Header() http.Header { return c.ResponseWriter.Header() }

// Write implements the interface ResponseWriter.
func (c *Context) Write(p []byte) (int, error) { return c.ResponseWriter.Write(p) }

// ---------------------------------------------------------------------------
// Common
// ---------------------------------------------------------------------------

// IsWebSocket reports whether the request is websocket.
func (c *Context) IsWebSocket() bool { return header.IsWebSocket(c.Request) }

// ContentType returns the Content-Type of the request without the charset.
func (c *Context) ContentType() string { return header.ContentType(c.Request.Header) }

// Charset returns the charset of the request content.
//
// Return "" if there is no charset.
func (c *Context) Charset() string { return header.Charset(c.Request.Header) }

// ---------------------------------------------------------------------------
// Request Query
// ---------------------------------------------------------------------------

// ParseQuery parses the query parameters, caches and returns the parsed query.
func (c *Context) ParseQuery() (query url.Values, err error) {
	if c.Query == nil {
		c.Query, err = url.ParseQuery(c.Request.URL.RawQuery)
	}
	query = c.Query
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

// ---------------------------------------------------------------------------
// Request Cookie
// ---------------------------------------------------------------------------

// GetCookies returns the HTTP cookies sent with the request.
func (c *Context) GetCookies() []*http.Cookie {
	if c.Cookies == nil {
		c.Cookies = c.Request.Cookies()
	}
	return c.Cookies
}

// GetCookie returns the named cookie provided in the request.
//
// Return nil if no the cookie named name.
func (c *Context) GetCookie(name string) *http.Cookie {
	cookies := c.GetCookies()
	for i, _len := 0, len(cookies); i < _len; i++ {
		if cookies[i].Name == name {
			return cookies[i]
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Response
// ---------------------------------------------------------------------------

// SetContentType sets the response header "Content-Type" to ct,
//
// If ct is "", do nothing.
func (c *Context) SetContentType(ct string) {
	header.SetContentType(c.ResponseWriter.Header(), ct)
}

// Redirect redirects the request to a provided URL with status code.
func (c *Context) Redirect(code int, toURL string) {
	if code < 300 || code >= 400 {
		panic(fmt.Errorf("invalid the redirect status code '%d'", code))
	}

	c.ResponseWriter.Header().Set(header.HeaderLocation, toURL)
	c.WriteHeader(code)
}

// Blob sends a blob response with the status code and the content type.
func (c *Context) Blob(code int, contentType string, data []byte) (err error) {
	c.SetContentType(contentType)
	c.WriteHeader(code)
	_, err = c.Write(data)
	return
}

// BlobText sends a string blob response with the status code and the content type.
func (c *Context) BlobText(code int, contentType string,
	format string, args ...interface{}) (err error) {
	c.SetContentType(contentType)
	c.WriteHeader(code)

	if len(args) > 0 {
		_, err = fmt.Fprintf(c.ResponseWriter, format, args...)
	} else {
		_, err = io.WriteString(c.ResponseWriter, format)
	}
	return
}

// Text sends a string response with the status code.
func (c *Context) Text(code int, format string, args ...interface{}) error {
	return c.BlobText(code, header.MIMETextPlainCharsetUTF8, format, args...)
}

// HTML sends a HTML response with the status code.
func (c *Context) HTML(code int, format string, args ...interface{}) error {
	return c.BlobText(code, header.MIMETextHTMLCharsetUTF8, format, args...)
}

// JSON sends a JSON response with the status code.
func (c *Context) JSON(code int, v interface{}) (err error) {
	buf := getBuilder()
	if err = json.NewEncoder(buf).Encode(v); err == nil {
		c.SetContentType(header.MIMEApplicationJSONCharsetUTF8)
		c.WriteHeader(code)
		_, err = buf.WriteTo(c.ResponseWriter)
	}
	putBuilder(buf)
	return
}

// XML sends a XML response with the status code.
func (c *Context) XML(code int, v interface{}) (err error) {
	buf := getBuilder()
	buf.WriteString(xml.Header)
	if err = xml.NewEncoder(buf).Encode(v); err == nil {
		c.SetContentType(header.MIMEApplicationXMLCharsetUTF8)
		c.WriteHeader(code)
		_, err = buf.WriteTo(c.ResponseWriter)
	}
	putBuilder(buf)
	return
}

// Stream sends a streaming response with the status code and the content type.
//
// If contentType is empty, Content-Type is ignored.
func (c *Context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.SetContentType(contentType)
	c.WriteHeader(code)
	_, err = io.CopyBuffer(c.ResponseWriter, r, make([]byte, 2048))
	return
}
