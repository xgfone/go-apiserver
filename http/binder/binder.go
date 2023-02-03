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

// Package binder is used to bind a value to the http request.
package binder

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/http/herrors"
	"github.com/xgfone/go-apiserver/tools/structfield"
)

// Predefine some binder to bind the body, query and header of the request.
var (
	DefaultMuxBinder = NewMuxBinder()

	DefaultQueryBinder Binder = BinderFunc(func(dst interface{}, r *http.Request) error {
		return helper.BindStructFromURLValues(dst, "query", r.URL.Query())
	})

	DefaultHeaderBinder Binder = BinderFunc(func(dst interface{}, r *http.Request) error {
		return helper.BindStructFromURLValues(dst, "header", url.Values(r.Header))
	})

	DefaultValidateFunc = func(v interface{}, r *http.Request) error {
		return structfield.Reflect(r, v)
	}

	BodyBinder   Binder = WrapBinder(DefaultMuxBinder, DefaultValidateFunc)
	QueryBinder  Binder = WrapBinder(DefaultQueryBinder, DefaultValidateFunc)
	HeaderBinder Binder = WrapBinder(DefaultHeaderBinder, DefaultValidateFunc)
)

func init() {
	DefaultMuxBinder.Add(header.MIMEApplicationXML, XMLBinder())
	DefaultMuxBinder.Add(header.MIMEApplicationJSON, JSONBinder())
	DefaultMuxBinder.Add(header.MIMEApplicationForm, FormBinder(10<<20))
	DefaultMuxBinder.Add(header.MIMEMultipartForm, FormBinder(10<<20))
}

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
		switch ct := header.ContentType(r.Header); ct {
		case header.MIMEMultipartForm:
			err = r.ParseMultipartForm(maxMemory)

		case header.MIMEApplicationForm:
			err = r.ParseForm()

		default:
			return fmt.Errorf("unsupported content-type '%s'", ct)
		}

		if err == nil {
			var fhs map[string][]*multipart.FileHeader
			if r.MultipartForm != nil {
				fhs = r.MultipartForm.File
			}

			err = helper.BindStruct(v, "form", func(name string) interface{} {
				if values, ok := r.Form[name]; ok {
					return values
				}
				if values, ok := fhs[name]; ok {
					return values
				}
				return nil
			})
		}

		return
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

	return herrors.ErrUnsupportedMediaType.WithMsg("not support Content-Type '%s'", ct)
}

// WrapBinder wraps the binder and returns a new one that goes to handle
// the result after binding the request.
func WrapBinder(binder Binder, nextHandler func(interface{}, *http.Request) error) Binder {
	return BinderFunc(func(dst interface{}, req *http.Request) (err error) {
		if err = binder.Bind(dst, req); err == nil {
			err = nextHandler(dst, req)
		}
		return
	})
}
