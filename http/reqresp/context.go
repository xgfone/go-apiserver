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
	"github.com/xgfone/go-apiserver/http/herrors"
	"github.com/xgfone/go-apiserver/http/render"
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
		return &Context{Data: make(map[string]interface{}, dataCap)}
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

	c.Request = req.WithContext(context.WithValue(req.Context(), reqParam(255), c))
	return c.Request
}

// GetContext returns a *Context, which (1) checks whether http.ResponseWriter
// has implemented the interface ContentGetter, or (2) extracts *Context from
// *http.Request.
//
// If the request context does not exist, reutrn nil.
func GetContext(w http.ResponseWriter, r *http.Request) *Context {
	switch c := w.(type) {
	case *Context:
		return c

	case ContextGetter:
		return c.GetContext(w, r)

	default:
		ctx, _ := r.Context().Value(reqParam(255)).(*Context)
		return ctx
	}
}

func handleContextResult(c *Context) {
	if !c.WroteHeader() {
		switch e := c.Err.(type) {
		case nil:
			c.WriteHeader(200)
		case herrors.Error:
			c.BlobText(e.Code, e.CT, c.Err.Error())
		default:
			c.Text(500, c.Err.Error())
		}
	}
}

// Handler is a handler to handle the http request.
type Handler func(*Context)

// ServeHTTP implements the interface http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := GetContext(w, r)
	if c == nil {
		c = DefaultContextAllocator.Acquire()
		c.ResponseWriter = NewResponseWriter(w)
		c.Request = r
		defer DefaultContextAllocator.Release(c)
	}
	h(c)
	handleContextResult(c)
}

// HandlerWithError is a handler to handle the http request with the error.
type HandlerWithError func(*Context) error

// ServeHTTP implements the interface http.Handler.
func (h HandlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := GetContext(w, r)
	if c == nil {
		c = DefaultContextAllocator.Acquire()
		c.ResponseWriter = NewResponseWriter(w)
		c.Request = r
		defer DefaultContextAllocator.Release(c)
	}
	c.Err = h(c)
	handleContextResult(c)
}

var _ ResponseWriter = &Context{}

// ContextGetter is used to get the Context from the http request or response.
type ContextGetter interface {
	// Return the Context from the http request or response.
	// If the Context does not exist, return nil.
	GetContext(http.ResponseWriter, *http.Request) *Context
}

// Context is used to represents the context information of the request.
type Context struct {
	ResponseWriter
	*http.Request

	// The context information, which will be reset to ZERO after finishing
	// to handle the request.
	Err  error                  // Be used to save the context error
	Reg1 interface{}            // The register to save the temporary context value.
	Reg2 interface{}            // The register to save the temporary context value.
	Reg3 interface{}            // The register to save the temporary context value.
	Data map[string]interface{} // A set of any key-value data

	// Render the content to the client.
	//
	// If nil, use render.DefaultRenderer instead.
	Renderer render.Renderer

	// Bind the value to the request body
	//
	// If nil, use binder.BodyBinder instead.
	BodyBinder binder.Binder

	// Bind the value to the request query.
	//
	// If nil, use binder.QueryBinder instead.
	QueryBinder binder.Binder

	// Bind the value to the request header.
	//
	// If nil, use binder.HeaderBinder instead.
	HeaderBinder binder.Binder

	// Query and Cookies are used to cache the parsed request query and cookies.
	Cookies []*http.Cookie
	Query   url.Values
}

// NewContext returns a new Context.
func NewContext(dataCapSize int) *Context {
	return &Context{Data: make(map[string]interface{}, dataCapSize)}
}

// GetContext implements the interface ContextGetter which returns itself.
func (c *Context) GetContext(http.ResponseWriter, *http.Request) *Context {
	return c
}

// Reset resets the context itself.
func (c *Context) Reset() {
	// Clean the datas.
	if len(c.Data) > 0 {
		for key := range c.Data {
			delete(c.Data, key)
		}
	}

	*c = Context{
		Data:         c.Data,
		Renderer:     c.Renderer,
		BodyBinder:   c.BodyBinder,
		QueryBinder:  c.QueryBinder,
		HeaderBinder: c.HeaderBinder,
	}
}

// BindBody extracts the data from the request body and assigns it to v.
func (c *Context) BindBody(v interface{}) (err error) {
	if c.BodyBinder == nil {
		err = binder.BodyBinder.Bind(v, c.Request)
	} else {
		err = c.BodyBinder.Bind(v, c.Request)
	}
	return
}

// BindQuery extracts the data from the request query and assigns it to v.
func (c *Context) BindQuery(v interface{}) (err error) {
	if c.QueryBinder == nil {
		err = binder.QueryBinder.Bind(v, c.Request)
	} else {
		err = c.QueryBinder.Bind(v, c.Request)
	}
	return
}

// BindHeader extracts the data from the request header and assigns it to v.
func (c *Context) BindHeader(v interface{}) (err error) {
	if c.HeaderBinder == nil {
		err = binder.HeaderBinder.Bind(v, c.Request)
	} else {
		err = c.HeaderBinder.Bind(v, c.Request)
	}
	return
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
// Data
// ---------------------------------------------------------------------------

// GetDataString returns the value as the string by the key from the field Data.
//
// If the key does not exist, return "".
func (c *Context) GetDataString(key string) string {
	if value, ok := c.Data[key]; ok {
		return value.(string)
	}
	return ""
}

// GetData returns the value by the key from the field Data.
//
// If the key does not exist, return nil.
func (c *Context) GetData(key string) interface{} {
	return c.Data[key]
}

// SetData sets the value with the key into the field Data.
//
// If value is nil, delete the key from the field Data.
func (c *Context) SetData(key string, value interface{}) {
	if value == nil {
		delete(c.Data, key)
	} else {
		c.Data[key] = value
	}
}

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

// Render renders the response with the name and the the data.
func (c *Context) Render(code int, name string, data interface{}) (err error) {
	if c.Renderer == nil {
		err = render.DefaultRenderer.Render(c.ResponseWriter, code, name, data)
	} else {
		err = c.Renderer.Render(c.ResponseWriter, code, name, data)
	}
	return
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
