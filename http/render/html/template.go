// Copyright 2022 xgfone
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

// Package html provides a template renderer, which is based on the stdlib
// "html/template", to render the HTML.
package html

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/http/header"
)

// NewTemplate returns a new template renderer to render the html.
func NewTemplate(loader Loader) *Template {
	r := &Template{loader: loader}
	r.tmpl.Store(template.New("__DEFAULT_HTML_TEMPLATE__"))
	r.bufs = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
	return r
}

// Template is used to render the html content from the templates.
type Template struct {
	loader Loader
	debug  int32
	right  string
	left   string
	funcs  []template.FuncMap

	tmpl atomic.Value // *template.Template
	bufs sync.Pool
	load sync.Once
}

// Debug sets the debug model and returns itself, which is thread-safe.
//
// If debug is true, it will reload all the templates automatically each time
// the template is rendered, which should be only used in the development.
func (r *Template) Debug(debug bool) *Template {
	if debug {
		atomic.StoreInt32(&r.debug, 1)
	} else {
		atomic.StoreInt32(&r.debug, 0)
	}
	return r
}

// Delims resets the left and right delimiter.
//
// The default delimiters are "{{" and "}}".
//
// Notice: it must be set before rendering the html template.
func (r *Template) Delims(left, right string) *Template {
	r.left, r.right = left, right
	return r
}

// Funcs appends the FuncMap.
//
// Notice: it must be set before rendering the html template.
func (r *Template) Funcs(funcs template.FuncMap) *Template {
	r.funcs = append(r.funcs, funcs)
	return r
}

// Reload reloads all the templates, which is thread-safe.
func (r *Template) Reload() error { return r.reload() }

func (r *Template) reload() error {
	files, err := r.loader.LoadAll()
	if err != nil {
		return err
	}

	tmpl := template.New("__DEFAULT_HTML_TEMPLATE__")
	tmpl.Delims(r.left, r.right)
	for _, file := range files {
		t := tmpl.New(file.Name())
		for _, funcs := range r.funcs {
			t.Funcs(funcs)
		}

		if _, err = t.Parse(file.Data()); err != nil {
			return err
		}
	}

	r.tmpl.Store(tmpl)
	return nil
}

// Render implements the interface render.Renderer to render the html content
// from the template files by the given template name.
func (r *Template) Render(w http.ResponseWriter, name string, code int, data interface{}) (err error) {
	if atomic.LoadInt32(&r.debug) == 1 {
		if err = r.reload(); err != nil {
			return
		}
	} else {
		r.load.Do(func() { err = r.reload() })
		if err != nil {
			return
		}
	}

	buf := r.bufs.Get().(*bytes.Buffer)
	if err = r.execute(buf, name, data); err == nil {
		header.SetContentType(w.Header(), header.MIMETextHTMLCharsetUTF8)
		w.WriteHeader(code)
		_, err = buf.WriteTo(w)
	}
	buf.Reset()
	r.bufs.Put(buf)

	return
}

func (r *Template) execute(w io.Writer, name string, data interface{}) error {
	return r.tmpl.Load().(*template.Template).ExecuteTemplate(w, name, data)
}
