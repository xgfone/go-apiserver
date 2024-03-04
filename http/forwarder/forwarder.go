// Copyright 2024 xgfone
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

// Package forwarder provides a http request forwarder to forward a request
// to another host.
package forwarder

import (
	"io"
	"net/http"
)

// DefaultForwarder is the default request forwarder.
var DefaultForwarder = NewForwarder("")

// Forward is equal to DefaultForwarder.Forward(w, r, host).
func Forward(w http.ResponseWriter, r *http.Request, host string) error {
	return DefaultForwarder.Forward(w, r, host)
}

// Forwarder is used to forwards a request to another host.
type Forwarder struct {
	Scheme string
	Host   string

	Client   *http.Client
	Request  func(*http.Request) *http.Request
	Response func(http.ResponseWriter, *http.Response) error
}

// NewForwarder returns a new forwarder which forwards a request to the host.
func NewForwarder(host string) *Forwarder {
	return &Forwarder{Host: host}
}

// Forward forwards the request to the host and copies the response to the client.
func (f *Forwarder) Forward(w http.ResponseWriter, r *http.Request, host string) (err error) {
	req := r.Clone(r.Context())
	req.RequestURI = ""          // Pretend to be a client request.
	req.Header.Del("Connection") // Enable the keepalive
	req.Close = false            // Enable the keepalive

	if f.Scheme == "" {
		req.URL.Scheme = "http"
	} else {
		req.URL.Scheme = f.Scheme
	}

	if host == "" {
		req.URL.Host = f.Host
	} else {
		req.URL.Host = host
	}

	if f.Request != nil {
		req = f.Request(req)
	}

	var resp *http.Response
	if f.Client == nil {
		resp, err = http.DefaultClient.Do(req)
	} else {
		resp, err = f.Client.Do(req)
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return
	}

	if f.Response == nil {
		err = CopyResponse(w, resp)
	} else {
		err = f.Response(w, resp)
	}

	return
}

// CopyResponse copyies the response to the request client.
func CopyResponse(w http.ResponseWriter, resp *http.Response) (err error) {
	// Copy response header
	header := w.Header()
	for k, vs := range resp.Header {
		switch {
		case len(vs) == 0:
		case len(vs) == 1 && vs[0] == "":

		case k == "Host":
		case k == "Connection":

		default:
			header[k] = vs
		}
	}

	// Copy response body
	w.WriteHeader(resp.StatusCode)
	_, err = io.CopyBuffer(w, resp.Body, make([]byte, 1024))
	return
}
