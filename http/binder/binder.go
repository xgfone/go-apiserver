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

// Package binder is used to bind a value to the http request.
package binder

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/http/herrors"
)

// Binder is used to bind the data to the http request.
type Binder interface {
	// Bind parses the data from http.Request to dst.
	//
	// Notice: dst must be a non-nil pointer.
	Bind(dst interface{}, req *http.Request) error
}

// BinderFunc is a function type implementing the interface Binder.
type BinderFunc func(dst interface{}, req *http.Request) error

// Bind implements the interface Binder.
func (f BinderFunc) Bind(dst interface{}, req *http.Request) error {
	return f(dst, req)
}

// BindQuery binds the data to the url query, which supports the struct tag "query".
func BindQuery(data interface{}, query url.Values) error {
	return BindURLValues(data, query, "query")
}

// JSONBinder returns a binder to bind the data to the request body as JSON.
func JSONBinder() Binder {
	return BinderFunc(func(v interface{}, r *http.Request) (err error) {
		if r.ContentLength > 0 {
			err = json.NewDecoder(r.Body).Decode(v)
		}
		return
	})
}

// XMLBinder returns a binder to bind the data to the request body as XML.
func XMLBinder() Binder {
	return BinderFunc(func(v interface{}, r *http.Request) (err error) {
		if r.ContentLength > 0 {
			err = xml.NewDecoder(r.Body).Decode(v)
		}
		return
	})
}

// FormBinder returns a binder to bind the data to the request body as Form,
// which supports the struct tag "form".
func FormBinder(maxMemory int64) Binder {
	return BinderFunc(func(v interface{}, r *http.Request) (err error) {
		ct := r.Header.Get("Content-Type")

		if strings.HasPrefix(ct, header.MIMEMultipartForm) {
			if err = r.ParseMultipartForm(maxMemory); err != nil {
				return
			}
		} else if err = r.ParseForm(); err != nil {
			return err
		}

		return BindURLValues(v, r.Form, "form")
	})
}

// MuxBinder is a multiplexer for kinds of Binders based on the request header
// "Content-Type".
type MuxBinder struct {
	binders map[string]Binder
}

// NewMuxBinder returns a new MuxBinder.
func NewMuxBinder() *MuxBinder {
	return &MuxBinder{binders: make(map[string]Binder, 8)}
}

// Add adds a binder to bind the content for the header "Content-Type".
func (mb *MuxBinder) Add(contentType string, binder Binder) {
	mb.binders[contentType] = binder
}

// Get returns the corresponding binder by the header "Content-Type".
//
// Return nil if not found.
func (mb *MuxBinder) Get(contentType string) Binder {
	return mb.binders[contentType]
}

// Del removes the corresponding binder by the header "Content-Type".
func (mb *MuxBinder) Del(contentType string) {
	delete(mb.binders, contentType)
}

// Bind implements the interface Binder, which looks up the registered binder
// by the request header "Content-Type" and calls it to bind the value dst
// to req.
func (mb *MuxBinder) Bind(dst interface{}, req *http.Request) error {
	ct := header.ContentType(req.Header)
	if ct == "" {
		return herrors.ErrMissingContentType
	}

	if binder := mb.Get(ct); binder != nil {
		return binder.Bind(dst, req)
	}

	return herrors.ErrUnsupportedMediaType.Newf("not support Content-Type '%s'", ct)
}

// DefaultValidateBinder is a binder with the validator and the default value setter.
type DefaultValidateBinder struct {
	Binder

	// SetDefault is used to set the data to the default if it is ZERO.
	//
	// If data is a struct, set the fields of the struct to the default if ZERO.
	//
	// Default: nil
	SetDefault func(data interface{}) error

	// Validate is used to validate whether data is valid.
	//
	// Default: nil
	Validate func(data interface{}) error
}

// Bind implements the interface Binder, and set the default value
// and validate whether the value is valid.
func (b *DefaultValidateBinder) Bind(v interface{}, r *http.Request) (err error) {
	if err = b.Binder.Bind(v, r); err != nil {
		return
	}

	if b.SetDefault != nil {
		if err = b.SetDefault(v); err != nil {
			return
		}
	}

	if b.Validate != nil {
		err = b.Validate(v)
	}

	return
}
