// Copyright 2021~2023 xgfone
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
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/http/render"
	"github.com/xgfone/go-apiserver/internal/errors2"
	"github.com/xgfone/go-apiserver/result"
	"github.com/xgfone/go-binder"
	"github.com/xgfone/go-cast"
	"golang.org/x/exp/maps"
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

// NewContextAllocator is equal to NewContextAllocatorWithResponseHandle(dataCap, nil).
func NewContextAllocator(dataCap int) ContextAllocator {
	return NewContextAllocatorWithResponseHandler(dataCap, nil)
}

// NewContextAllocatorWithResponseHandler returns a new ContextAllocator,
// which acquires a request context from the pool and releases the request
// context into the pool.
//
// Notice: if Context.Any has implemented the interface { Reset() },
// it will be called when releasing the request context.
func NewContextAllocatorWithResponseHandler(dataCap int, handler ResponseHandler) ContextAllocator {
	return NewContextAllocatorWithUpdater(dataCap, func(c *Context) { c.ResponseHandler = handler })
}

func NewContextAllocatorWithUpdater(dataCap int, update func(*Context)) ContextAllocator {
	var alloc contextAllocator
	if update == nil {
		alloc.ctxPool.New = func() interface{} {
			return &Context{Data: make(map[string]interface{}, dataCap)}
		}
	} else {
		alloc.ctxPool.New = func() interface{} {
			c := &Context{Data: make(map[string]interface{}, dataCap)}
			update(c)
			return c
		}
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
	return req.WithContext(SetContextIntoCtx(req.Context(), c))
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
		return GetContextFromCtx(r.Context())
	}
}

// SetContextIntoCtx sets *Context into context.Context and returns the new one.
func SetContextIntoCtx(ctx context.Context, c *Context) context.Context {
	return context.WithValue(ctx, reqParam(255), c)
}

// GetContextFromCtx returns a *Context from the context.
//
// If the request context does not exist, reutrn nil.
func GetContextFromCtx(ctx context.Context) *Context {
	c, _ := ctx.Value(reqParam(255)).(*Context)
	return c
}

func handleContextResult(c *Context) {
	switch e := c.Err.(type) {
	case nil:
		c.WriteHeader(200)
	case http.Handler:
		e.ServeHTTP(c.ResponseWriter, c.Request)
	default:
		c.Text(500, c.Err.Error())
	}
}

func handleDefault(c *Context) {
	if !c.WroteHeader() {
		if c.DefaultHandler != nil {
			c.DefaultHandler(c)
		} else if DefaultHandler != nil {
			DefaultHandler(c)
		} else {
			handleContextResult(c)
		}
	}
}

// DefaultHandler is used to handle the response if not wrote the response header.
var DefaultHandler func(*Context) = handleContextResult

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
	handleDefault(c)
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
	c.UpdateError(h(c))
	handleDefault(c)
}

var (
	_ ResponseWriter = new(Context)
	_ ContextGetter  = new(Context)
)

// ContextGetter is used to get the Context from the http request or response.
type ContextGetter interface {
	// Return the Context from the http request or response.
	// If the Context does not exist, return nil.
	GetContext(http.ResponseWriter, *http.Request) *Context
}

// UpdateContextError is a global function to update the context error,
// which will be used by Context.UpdateError.
var UpdateContextError func(c *Context, err error)

// DefaultResponseHandler is used by Context
// when Context.ResponseHandler is not set.
var DefaultResponseHandler ResponseHandler

// ResponseHandler is used to handle the response.
type ResponseHandler func(*Context, result.Response) error

// Context is used to represents the context information of the request.
type Context struct {
	ResponseWriter
	*http.Request

	// The context information, which will be reset to ZERO after finishing
	// to handle the request.
	Err  error       // Be used to save the context error
	Reg1 interface{} // The register to save the temporary context value.
	Reg2 interface{} // The register to save the temporary context value.
	Reg3 interface{} // The register to save the temporary context value.
	// As a general rule, the keys starting with "_" are private.
	Data map[string]interface{} // A set of any key-value pairs

	// The extra context information, which may be used by some middlewares
	// or services, such as the action router.
	Action  string
	Handler http.Handler

	// Render the content to the client.
	//
	// If nil, use render.DefaultRenderer instead.
	Renderer render.Renderer

	// Bind the value to the request body
	//
	// If nil, use binder.BodyDecoder instead.
	BodyDecoder binder.Decoder

	// Bind the value to the request query.
	//
	// If nil, use binder.QueryDecoder instead.
	QueryDecoder binder.Decoder

	// Bind the value to the request header.
	//
	// If nil, use binder.HeaderDecoder instead.
	HeaderDecoder binder.Decoder

	// HandleResponse is used to wrap the response and handle it by itself.
	ResponseHandler ResponseHandler

	// DefaultHandler is used to handle the response if not wrote the response header.
	DefaultHandler func(*Context)

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

// GetRequest returns the http.Request.
func (c *Context) GetRequest() *http.Request { return c.Request }

// GetResponse returns the http.ResponseWriter.
func (c *Context) GetResponse() http.ResponseWriter { return c.ResponseWriter }

// UpdateError updates the context error.
func (c *Context) UpdateError(err error) {
	if UpdateContextError != nil {
		UpdateContextError(c, err)
	} else if err != nil {
		if c.Err == nil {
			c.Err = err
		} else {
			c.Err = errors2.Join(c.Err, err)
		}
	}
}

// Reset resets the context itself.
func (c *Context) Reset() {
	maps.Clear(c.Data)
	*c = Context{
		Data:            c.Data,
		Renderer:        c.Renderer,
		BodyDecoder:     c.BodyDecoder,
		QueryDecoder:    c.QueryDecoder,
		HeaderDecoder:   c.HeaderDecoder,
		ResponseHandler: c.ResponseHandler,
	}
}

// BindBody extracts the data from the request body and assigns it to v.
func (c *Context) BindBody(v interface{}) (err error) {
	if c.BodyDecoder == nil {
		err = binder.BodyDecoder.Decode(v, c.Request)
	} else {
		err = c.BodyDecoder.Decode(v, c.Request)
	}
	return
}

// BindQuery extracts the data from the request query and assigns it to v.
func (c *Context) BindQuery(v interface{}) (err error) {
	if c.QueryDecoder == nil {
		err = binder.QueryDecoder.Decode(v, c.Request)
	} else {
		err = c.QueryDecoder.Decode(v, c.Request)
	}
	return
}

// BindHeader extracts the data from the request header and assigns it to v.
func (c *Context) BindHeader(v interface{}) (err error) {
	if c.HeaderDecoder == nil {
		err = binder.HeaderDecoder.Decode(v, c.Request)
	} else {
		err = c.HeaderDecoder.Decode(v, c.Request)
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

// LocalAddr returns the local address of the request connection.
func (c *Context) LocalAddr() net.Addr {
	return c.Request.Context().Value(http.LocalAddrContextKey).(net.Addr)
}

// RequestID returns the request header "X-Request-Id".
func (c *Context) RequestID() string { return c.Request.Header.Get(header.HeaderXRequestID) }

// IsWebSocket reports whether the request is websocket.
func (c *Context) IsWebSocket() bool { return header.IsWebSocket(c.Request) }

// ContentType returns the Content-Type of the request without the charset.
func (c *Context) ContentType() string { return header.ContentType(c.Request.Header) }

// Charset returns the charset of the request content.
//
// Return "" if there is no charset.
func (c *Context) Charset() string { return header.Charset(c.Request.Header) }

// Accept returns the accepted Content-Type list from the request header
// "Accept", which are sorted by the q-factor weight from high to low.
//
// If there is no the request header "Accept", return nil.
func (c *Context) Accept() []string { return header.Accept(c.Request.Header) }

// Scheme returns the HTTP protocol scheme, `http` or `https`.
func (c *Context) Scheme() string {
	if c.Request.TLS != nil {
		return "https"
	}
	return header.Scheme(c.Request.Header)
}

// ---------------------------------------------------------------------------
// Data
// ---------------------------------------------------------------------------

// GetDataInt64 returns the value as int64 by the key from the field Data.
//
// If the key does not exist and required is false, return (0, nil).
func (c *Context) GetDataInt64(key string, required bool) (value int64, err error) {
	if v, exist := c.Data[key]; exist {
		value, err = cast.ToInt64(v)
	} else if required {
		err = fmt.Errorf("missing %s", key)
	}
	return
}

// GetDataUint64 returns the value as uint64 by the key from the field Data.
//
// If the key does not exist and required is false, return (0, nil).
func (c *Context) GetDataUint64(key string, required bool) (value uint64, err error) {
	if v, exist := c.Data[key]; exist {
		value, err = cast.ToUint64(v)
	} else if required {
		err = fmt.Errorf("missing %s", key)
	}
	return
}

// GetDataString returns the value as string by the key from the field Data.
//
// If the key does not exist, return "".
func (c *Context) GetDataString(key string) string {
	if value, ok := c.Data[key]; ok {
		return cast.Must(cast.ToString(value))
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

// GetQueryInt64 returns the value as int64 by the key from the field Data.
//
// If the key does not exist and required is false, return (0, nil).
func (c *Context) GetQueryInt64(key string, required bool) (value int64, err error) {
	if vs, exist := c.GetQueries()[key]; exist {
		switch len(vs) {
		case 0:
		case 1:
			value, err = cast.ToInt64(vs[0])
		default:
			err = fmt.Errorf("too query values for %s", key)
		}
	} else if required {
		err = fmt.Errorf("missing %s", key)
	}
	return
}

// GetQueryUint64 returns the value as uint64 by the key from the field Data.
//
// If the key does not exist and required is false, return (0, nil).
func (c *Context) GetQueryUint64(key string, required bool) (value uint64, err error) {
	if vs, exist := c.GetQueries()[key]; exist {
		switch len(vs) {
		case 0:
		case 1:
			value, err = cast.ToUint64(vs[0])
		default:
			err = fmt.Errorf("too query values for %s", key)
		}
	} else if required {
		err = fmt.Errorf("missing %s", key)
	}
	return
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

// SetConnectionClose sets the response header "Connection: close"
// to tell the server to close the connection.
func (c *Context) SetConnectionClose() {
	c.ResponseWriter.Header().Set(header.HeaderConnection, "close")
}

// SetContentType sets the response header "Content-Type" to ct,
//
// If ct is "", do nothing.
func (c *Context) SetContentType(ct string) {
	header.SetContentType(c.ResponseWriter.Header(), ct)
}

func (c *Context) sendfile(name, path, dtype string) (err error) {
	if name == "" {
		name = filepath.Base(path)
	}

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return
	} else if stat.IsDir() {
		return fmt.Errorf("file '%s' is a directory", path)
	}

	params := map[string]string{"filename": name}
	disposition := mime.FormatMediaType(dtype, params)
	c.ResponseWriter.Header().Set(header.HeaderContentDisposition, disposition)

	http.ServeContent(c.ResponseWriter, c.Request, stat.Name(), stat.ModTime(), file)
	return
}

// Attachment sends a file as attachment.
//
// If filename is "", it will use the base name of the filepath instead.
// And if the file does not exist, it returns os.ErrNotExist.
func (c *Context) Attachment(filename, filepath string) error {
	if filepath == "" {
		panic("Context.Attachment: filepath must not be empty")
	}
	return c.sendfile(filename, filepath, "attachment")
}

// Inline sends a file as inline.
//
// If filename is "", it will use the base name of the filepath instead.
// And if the file does not exist, it returns os.ErrNotExist.
func (c *Context) Inline(filename, filepath string) error {
	if filepath == "" {
		panic("Context.Inline: filepath must not be empty")
	}
	return c.sendfile(filename, filepath, "inline")
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

// Respond forwards the calling to c.ResponseHandler if it is set.
// Or, try to use DefaultResponseHandler instead if set.
// Or, it is equal to c.JSON(200, response).
func (c *Context) Respond(response result.Response) {
	if response.Error != nil {
		c.UpdateError(response.Error)
	}

	var err error
	if c.ResponseHandler != nil {
		err = c.ResponseHandler(c, response)
	} else if DefaultResponseHandler != nil {
		DefaultResponseHandler(c, response)
	} else {
		err = c.JSON(200, response)
	}

	c.UpdateError(err)
}

// Success is equal to c.Respond(result.Response{Data: data}).
func (c *Context) Success(data interface{}) {
	c.Respond(result.Response{Data: data})
}

// Failure is the same as c.Respond(result.Response{Error: err})
// if err is not nil. Or, it is equal to c.Success(nil).
func (c *Context) Failure(err error) {
	switch e := err.(type) {
	case nil, result.Error:
	case result.CodeError:
		err = e.CodeError()
	default:
		err = result.ErrInternalServerError.WithError(err)
	}
	c.Respond(result.Response{Error: err})
}

// Error sends the error as the response, and returns the sent error.
//
//	If err is nil, it is equal to c.WriteHeader(200).
//	If err implements http.Handler, it is equal to err.ServeHTTP(c.ResponseWriter, c.Request).
//	Or, it is equal to c.Text(500, err.Error()).
//
// Notice: herrors.Error has implements the interface http.Handler.
func (c *Context) Error(err error) error {
	c.UpdateError(err)
	switch e := err.(type) {
	case nil:
		c.WriteHeader(200)

	case http.Handler:
		e.ServeHTTP(c.ResponseWriter, c.Request)

	default:
		return c.Text(500, err.Error())
	}

	return nil
}

// Blob sends a blob response with the status code and the content type.
func (c *Context) Blob(code int, contentType string, data []byte) (err error) {
	c.SetContentType(contentType)
	c.WriteHeader(code)
	if len(data) > 0 {
		_, err = c.Write(data)
	}
	return
}

// BlobText sends a string blob response with the status code and the content type.
func (c *Context) BlobText(code int, contentType string,
	format string, args ...interface{}) (err error) {
	c.SetContentType(contentType)
	c.WriteHeader(code)

	if len(format) > 0 {
		if len(args) > 0 {
			_, err = fmt.Fprintf(c.ResponseWriter, format, args...)
		} else {
			_, err = io.WriteString(c.ResponseWriter, format)
		}
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
	if v == nil {
		c.SetContentType(header.MIMEApplicationJSONCharsetUTF8)
		c.WriteHeader(code)
		return
	}

	buf := getBuilder()
	if err = helper.EncodeJSON(buf, v); err == nil {
		c.SetContentType(header.MIMEApplicationJSONCharsetUTF8)
		c.WriteHeader(code)
		_, err = buf.WriteTo(c.ResponseWriter)
	}
	putBuilder(buf)
	return
}

// XML sends a XML response with the status code.
func (c *Context) XML(code int, v interface{}) (err error) {
	if v == nil {
		c.SetContentType(header.MIMEApplicationXMLCharsetUTF8)
		c.WriteHeader(code)
		return
	}

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
	_, err = io.CopyBuffer(c.ResponseWriter, r, make([]byte, 1024))
	return
}

// NoContent is the alias of WriteHeader.
func (c *Context) NoContent(code int) {
	c.WriteHeader(code)
}
