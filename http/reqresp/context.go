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
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-binder"
	"github.com/xgfone/go-cast"
)

type contextkey struct{ key uint8 }

var ctxkey = contextkey{key: 255}

// SetContext returns a new context.Context containing c.
func SetContext(ctx context.Context, c *Context) context.Context {
	return context.WithValue(ctx, ctxkey, c)
}

// GetContext returns a *Context from context.Context.
//
// If not exist, reutrn nil.
func GetContext(ctx context.Context) *Context {
	c, _ := ctx.Value(ctxkey).(*Context)
	return c
}

var ctxpool = &sync.Pool{New: func() any { return NewContext(4) }}

// AcquireContext acquires a context from the pool.
func AcquireContext() *Context { return ctxpool.Get().(*Context) }

// ReleaseContext releases the context to the pool.
func ReleaseContext(c *Context) { c.Reset(); ctxpool.Put(c) }

// Context is used to represents the context information of the request.
type Context struct {
	ResponseWriter
	*http.Request

	// As a general rule, the data keys starting with "_" are private.
	Data map[string]interface{} // A set of any key-value pairs
	Reg1 interface{}            // The register to save the temporary context value.
	Reg2 interface{}            // The register to save the temporary context value.
	Reg3 interface{}            // The register to save the temporary context value.
	Err  error                  // Used to save the context error.

	// The extra context information, which may be used by other service,
	// such as the action router.
	Version string
	Action  string

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

	// Query and Cookies are used to cache the parsed request query and cookies.
	Cookies []*http.Cookie
	Query   url.Values
}

// NewContext returns a new Context.
func NewContext(dataCapSize int) *Context {
	return &Context{Data: make(map[string]interface{}, dataCapSize)}
}

// Reset resets the context itself.
func (c *Context) Reset() {
	clear(c.Data)
	*c = Context{
		Data: c.Data,
		// Renderer:        c.Renderer,
		BodyDecoder:   c.BodyDecoder,
		QueryDecoder:  c.QueryDecoder,
		HeaderDecoder: c.HeaderDecoder,
		// DefaultHandler:  c.DefaultHandler,
		// ResponseHandler: c.ResponseHandler,
	}
}

// GetRequest returns the wrapped http.Request.
func (c *Context) GetRequest() *http.Request { return c.Request }

// GetResponse returns the wrapped http.ResponseWriter.
func (c *Context) GetResponse() http.ResponseWriter { return c.ResponseWriter }

// Header implements the interface http.ResponseWriter#Header.
func (c *Context) Header() http.Header { return c.ResponseWriter.Header() }

// Write implements the interface http.ResponseWriter#Write.
func (c *Context) Write(p []byte) (int, error) { return c.ResponseWriter.Write(p) }

// ---------------------------------------------------------------------------
// Binder
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Request Information
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
			value, err = strconv.ParseInt(vs[0], 10, 64)
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
			value, err = strconv.ParseUint(vs[0], 10, 64)
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

// AppendError appends the error err into c.Err.
func (c *Context) AppendError(err error) {
	if err != nil {
		if c.Err == nil {
			c.Err = err
		} else {
			c.Err = errors.Join(c.Err, err)
		}
	}
}

// SetConnectionClose sets the response header "Content-Disposition".
// For example,
//
//	Content-Disposition: inline
//	Content-Disposition: attachment
//	Content-Disposition: attachment; filename="filename.jpg"
//
// dtype must be either "inline" or "attachment". But, filename is optional.
func (c *Context) SetContentDisposition(dtype, filename string) {
	switch dtype {
	case "inline", "attachment":
	default:
		panic(fmt.Errorf("Context.SetContentDisposition: unknown dtype '%s'", dtype))
	}

	var disposition string
	if filename == "" {
		disposition = "Content-Disposition: " + dtype
	} else {
		params := map[string]string{"filename": filename}
		disposition = mime.FormatMediaType(dtype, params)
	}

	c.ResponseWriter.Header().Set(header.HeaderContentDisposition, disposition)
}

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

// NoContent is the alias of WriteHeader.
func (c *Context) NoContent(code int) { c.WriteHeader(code) }

// Redirect redirects the request to a provided URL with status code.
func (c *Context) Redirect(code int, toURL string) {
	if code < 300 || code >= 400 {
		panic(fmt.Errorf("invalid the redirect status code '%d'", code))
	}

	c.ResponseWriter.Header().Set(header.HeaderLocation, toURL)
	c.WriteHeader(code)
}

// Error sends the error as the response to the client
//
//	If err is nil, it is equal to c.WriteHeader(200).
//	If err implements http.Handler, it is equal to err.ServeHTTP(c.ResponseWriter, c.Request).
//	Or, it is equal to c.Text(500, err.Error()).
func (c *Context) Error(err error) {
	c.AppendError(c.Err)
	switch e := err.(type) {
	case nil:
		c.WriteHeader(200)

	case http.Handler:
		e.ServeHTTP(c.ResponseWriter, c.Request)

	default:
		c.Text(500, err.Error())
	}
}

// Blob sends a blob response with the status code and the content type.
func (c *Context) Blob(code int, contentType string, data []byte) {
	c.SetContentType(contentType)
	c.WriteHeader(code)
	if len(data) > 0 {
		_, err := c.Write(data)
		c.AppendError(err)
	}
}

// BlobText sends a string blob response with the status code and the content type.
func (c *Context) BlobText(code int, contentType string, format string, args ...interface{}) {
	c.SetContentType(contentType)
	c.WriteHeader(code)

	if len(format) > 0 {
		var err error
		if len(args) > 0 {
			_, err = fmt.Fprintf(c.ResponseWriter, format, args...)
		} else {
			_, err = io.WriteString(c.ResponseWriter, format)
		}
		c.AppendError(err)
	}
}

// Text sends a string response with the status code.
func (c *Context) Text(code int, format string, args ...interface{}) {
	c.BlobText(code, header.MIMETextPlainCharsetUTF8, format, args...)
}

// HTML sends a HTML response with the status code.
func (c *Context) HTML(code int, format string, args ...interface{}) {
	c.BlobText(code, header.MIMETextHTMLCharsetUTF8, format, args...)
}

// JSON sends a JSON response with the status code.
func (c *Context) JSON(code int, v interface{}) {
	if v == nil {
		c.WriteHeader(code)
		return
	}

	buf := getBuilder()
	if err := helper.EncodeJSON(buf, v); err == nil {
		c.SetContentType(header.MIMEApplicationJSONCharsetUTF8)
		c.WriteHeader(code)
		_, err := buf.WriteTo(c.ResponseWriter)
		c.AppendError(err)
	}
	putBuilder(buf)
}

// XML sends a XML response with the status code.
func (c *Context) XML(code int, v interface{}) {
	if v == nil {
		c.WriteHeader(code)
		return
	}

	buf := getBuilder()
	_, _ = buf.WriteString(xml.Header)
	if err := xml.NewEncoder(buf).Encode(v); err == nil {
		c.SetContentType(header.MIMEApplicationXMLCharsetUTF8)
		c.WriteHeader(code)
		_, err := buf.WriteTo(c.ResponseWriter)
		c.AppendError(err)
	}
	putBuilder(buf)
}

// Stream sends a streaming response with the status code and the content type.
//
// If contentType is empty, Content-Type is ignored.
func (c *Context) Stream(code int, contentType string, r io.Reader) {
	c.SetContentType(contentType)
	c.WriteHeader(code)
	_, err := io.CopyBuffer(c.ResponseWriter, r, make([]byte, 1024))
	c.AppendError(err)
}

// Attachment sends a file as attachment.
//
// If filename is "", it will use the base name of the filepath instead.
// And if the file does not exist, it returns os.ErrNotExist.
func (c *Context) Attachment(filename, filepath string) {
	if filepath == "" {
		panic("Context.Attachment: filepath must not be empty")
	}
	c.sendfile(filename, filepath, "attachment")
}

// Inline sends a file as inline.
//
// If filename is "", it will use the base name of the filepath instead.
// And if the file does not exist, it returns os.ErrNotExist.
func (c *Context) Inline(filename, filepath string) {
	if filepath == "" {
		panic("Context.Inline: filepath must not be empty")
	}
	c.sendfile(filename, filepath, "inline")
}

func (c *Context) sendfile(name, path, dtype string) {
	if name == "" {
		name = filepath.Base(path)
	}

	file, err := os.Open(path)
	if err != nil {
		c.AppendError(err)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return
	} else if stat.IsDir() {
		c.AppendError(fmt.Errorf("file '%s' is a directory", path))
		return
	}

	c.SetContentDisposition(dtype, name)
	http.ServeContent(c.ResponseWriter, c.Request, stat.Name(), stat.ModTime(), file)
}

/// ----------------------------------------------------------------------- ///

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
