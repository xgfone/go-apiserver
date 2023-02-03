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

package upstream

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	httpclient "github.com/xgfone/go-apiserver/http/client"
	"github.com/xgfone/go-apiserver/nets"
)

// ServerConfig is used to configure the http server.
type ServerConfig struct {
	// Required
	URL

	// Optional
	//
	// Default: URL.ID()
	ID string

	// Handle the request or response.
	//
	// Optional
	HTTPClient     httpclient.Getter // Default: use http.DefaultClient
	HandleRequest  func(*http.Client, *http.Request) (*http.Response, error)
	HandleResponse func(http.ResponseWriter, *http.Response) error

	// If DynamicWeight is configured, use it. Or use StaticWeight.
	//
	// Optional
	DynamicWeight func() int
	StaticWeight  int // Default: 1

	// GetServerStatus is used to check the server status dynamically.
	//
	// If nil, use ServerStatusOnline instead.
	GetServerStatus func(Server) ServerStatus
}

// NewServer returns a new server supporting the weight.
func NewServer(conf ServerConfig) (WeightedServer, error) {
	if conf.Hostname == "" && conf.IP == "" {
		return nil, fmt.Errorf("missing the hostname or ip")
	}
	if conf.Scheme == "" {
		conf.Scheme = "http"
	}
	if conf.StaticWeight <= 0 {
		conf.StaticWeight = 1
	}

	var addr string
	if conf.Port > 0 {
		if conf.IP != "" {
			addr = net.JoinHostPort(conf.IP, fmt.Sprint(conf.Port))
		} else {
			addr = net.JoinHostPort(conf.Hostname, fmt.Sprint(conf.Port))
		}
	} else {
		if conf.IP != "" {
			addr = conf.IP
		} else {
			addr = conf.Hostname
		}
	}

	id := conf.ID
	if id == "" {
		id = conf.URL.ID()
	}

	host := conf.Hostname
	if host != "" && conf.Port > 0 {
		host = net.JoinHostPort(conf.Hostname, strconv.FormatInt(int64(conf.Port), 10))
	}

	s := &httpServer{
		id:             id,
		addr:           addr,
		host:           host,
		conf:           conf,
		getWeight:      conf.DynamicWeight,
		getStatus:      conf.GetServerStatus,
		handleRequest:  conf.HandleRequest,
		handleResponse: conf.HandleResponse,
	}

	if s.getWeight == nil {
		s.getWeight = s.getStaticWeight
	}
	if s.getStatus == nil {
		s.getStatus = s.getStaticStatus
	}

	if s.handleRequest == nil {
		s.handleRequest = handleRequest
	}

	if s.handleResponse == nil {
		s.handleResponse = handleResponse
	}

	if _len := len(conf.Queries); _len > 0 {
		s.queries = make(url.Values, _len)
		for key, value := range conf.Queries {
			s.queries.Set(key, value)
		}
		s.rawQuery = s.queries.Encode()
	}

	if _len := len(conf.Headers); _len > 0 {
		s.headers = make(http.Header, _len)
		for key, value := range conf.Headers {
			s.headers.Set(key, value)
		}
	}

	return s, nil
}

var _ WeightedServer = &httpServer{}

type httpServer struct {
	id       string
	addr     string
	host     string
	conf     ServerConfig
	headers  http.Header
	queries  url.Values
	rawQuery string
	state    nets.RuntimeState

	getWeight      func() int
	getStatus      func(Server) ServerStatus
	handleRequest  func(*http.Client, *http.Request) (*http.Response, error)
	handleResponse func(http.ResponseWriter, *http.Response) error
}

func (s *httpServer) ID() string                      { return s.id }
func (s *httpServer) URL() URL                        { return s.conf.URL }
func (s *httpServer) Weight() int                     { return s.getWeight() }
func (s *httpServer) Status() ServerStatus            { return s.getStatus(s) }
func (s *httpServer) RuntimeState() nets.RuntimeState { return s.state.Clone() }
func (s *httpServer) HandleHTTP(w http.ResponseWriter, r *http.Request) (err error) {
	s.state.Inc()
	defer func() {
		s.state.Dec()
		if err == nil {
			s.state.IncSuccess()
		}
	}()

	r = r.Clone(r.Context())
	r.RequestURI = ""   // Pretend to be a client request.
	r.URL.Host = s.addr // Dial to the upstream backend server.
	r.URL.Scheme = s.conf.Scheme

	if s.conf.Method != "" {
		r.Method = s.conf.Method
	}

	if s.host != "" {
		// Override the header "Host"
		r.Host = s.host
	}

	if s.conf.Path != "" {
		r.URL.Path = s.conf.Path
	}

	if len(s.queries) > 0 {
		if r.URL.RawQuery == "" {
			r.URL.RawQuery = s.rawQuery
		} else if values := r.URL.Query(); len(values) == 0 {
			r.URL.RawQuery = s.rawQuery
		} else {
			for k, vs := range s.queries {
				values[k] = vs
			}
			r.URL.RawQuery = values.Encode()
		}
	}

	if len(s.headers) > 0 {
		if len(r.Header) == 0 {
			r.Header = s.headers.Clone()
		} else {
			for k, vs := range s.headers {
				r.Header[k] = vs
			}
		}
	}

	var resp *http.Response
	if s.conf.HTTPClient == nil {
		resp, err = s.handleRequest(http.DefaultClient, r)
	} else {
		resp, err = s.handleRequest(s.conf.HTTPClient.GetHTTPClient(), r)
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if err == nil {
		err = s.handleResponse(w, resp)
	}
	return err
}

func (s *httpServer) getStaticWeight() int                { return s.conf.StaticWeight }
func (s *httpServer) getStaticStatus(Server) ServerStatus { return ServerStatusOnline }

func handleRequest(c *http.Client, r *http.Request) (*http.Response, error) {
	return c.Do(r)
}

func handleResponse(w http.ResponseWriter, resp *http.Response) error {
	respHeader := w.Header()
	for k, vs := range resp.Header {
		respHeader[k] = vs
	}

	w.WriteHeader(resp.StatusCode)
	io.CopyBuffer(w, resp.Body, make([]byte, 1024))
	return nil
}

func (s *httpServer) Check(ctx context.Context, health URL) (err error) {
	if health.Method == "" {
		health.Method = http.MethodGet
	}

	if health.Scheme == "" {
		health.Scheme = s.conf.Scheme
	}

	if health.Hostname == "" {
		health.Hostname = s.conf.Hostname
	}

	health.IP = s.conf.IP
	if health.Port == 0 {
		health.Port = s.conf.Port
	}

	req, err := health.Request(ctx)
	if err != nil {
		return err
	}

	var resp *http.Response
	if s.conf.HTTPClient == nil {
		resp, err = http.DefaultClient.Do(req)
	} else {
		resp, err = s.conf.HTTPClient.GetHTTPClient().Do(req)
	}

	if resp != nil {
		resp.Body.Close()
	}

	if err == nil && resp.StatusCode >= 400 {
		err = fmt.Errorf("unexpected status code '%d'", resp.StatusCode)
	}

	return
}
