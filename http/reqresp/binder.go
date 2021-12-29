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
	"encoding/json"
	"encoding/xml"
	"net/http"

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/http/herrors"
)

// Binder is used to bind the request data to a struct.
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