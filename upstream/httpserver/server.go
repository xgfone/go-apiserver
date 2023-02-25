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

package httpserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"

	httpclient "github.com/xgfone/go-apiserver/http/client"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/nets"
	"github.com/xgfone/go-apiserver/upstream"
)

// Config is used to configure the http server.
type Config struct {
	// Hostname or IP is required, and others are optional.
	URL URL

	// CheckURL is used to check the http upstream server is healthy.
	//
	// Optional
	CheckURL URL

	// Optional
	//
	// Default: URL.ID()
	ID string

	// The extra information about the http upstream server.
	//
	// Optional
	Info interface{}

	// Handle the request or response.
	//
	// Optional
	HTTPClient     httpclient.Getter // Default: use http.DefaultClient
	HandleRequest  func(*http.Client, *http.Request) (*http.Response, error)
	HandleResponse func(http.ResponseWriter, *http.Response) error

	// If DynamicWeight is configured, use it. Or use StaticWeight.
	//
	// Optional
	DynamicWeight func(upstream.Server) int
	StaticWeight  int // Default: 1

	// GetServerStatus is used to check the server status dynamically.
	//
	// If nil, use ServerStatusOnline instead.
	GetServerStatus func(upstream.Server) upstream.ServerStatus
}

// NewServer returns a new upstream.WeightedServer based on HTTP.
func (c Config) NewServer() (upstream.WeightedServer, error) {
	s := &httpServer{}
	s.Update(c)
	return s, nil
}

type config struct {
	Config

	id   string
	addr string
	host string

	headers  http.Header
	queries  url.Values
	rawQuery string

	getClient      func() *http.Client
	getWeight      func(upstream.Server) int
	getStatus      func(upstream.Server) upstream.ServerStatus
	handleRequest  func(*http.Client, *http.Request) (*http.Response, error)
	handleResponse func(http.ResponseWriter, *http.Response) error
}

func newConfig(c Config) (conf config, err error) {
	conf.Config = c
	err = conf.init()
	return
}

func (c *config) init() error {
	c.URL = c.URL.Clone()
	c.CheckURL = c.CheckURL.Clone()

	if c.URL.Hostname == "" && c.URL.IP == "" {
		return fmt.Errorf("missing the hostname and ip")
	}
	if c.URL.Scheme == "" {
		c.URL.Scheme = "http"
	}
	if c.StaticWeight <= 0 {
		c.StaticWeight = 1
	}

	if c.URL.Port > 0 {
		if c.URL.IP != "" {
			c.addr = net.JoinHostPort(c.URL.IP, fmt.Sprint(c.URL.Port))
		} else {
			c.addr = net.JoinHostPort(c.URL.Hostname, fmt.Sprint(c.URL.Port))
		}
	} else {
		if c.URL.IP != "" {
			c.addr = c.URL.IP
		} else {
			c.addr = c.URL.Hostname
		}
	}

	c.id = c.ID
	if c.id == "" {
		c.id = c.URL.ID()
	}

	c.host = c.URL.Hostname
	if c.host != "" && c.URL.Port > 0 {
		c.host = net.JoinHostPort(c.URL.Hostname, strconv.FormatInt(int64(c.URL.Port), 10))
	}

	c.headers = map2httpheader(c.URL.Headers)
	c.queries = map2urlvalues(c.URL.Queries)
	c.rawQuery = c.queries.Encode()

	c.getWeight = c.DynamicWeight
	c.getStatus = c.GetServerStatus
	c.handleRequest = c.HandleRequest
	c.handleResponse = c.HandleResponse
	if c.getWeight == nil {
		c.getWeight = c.getStaticWeight
	}
	if c.getStatus == nil {
		c.getStatus = getStaticStatus
	}
	if c.handleRequest == nil {
		c.handleRequest = handleRequest
	}
	if c.handleResponse == nil {
		c.handleResponse = handleResponse
	}
	if c.HTTPClient == nil {
		c.getClient = getHTTPClient
	} else {
		c.getClient = c.HTTPClient.GetHTTPClient
	}

	if c.CheckURL.Method == "" {
		c.CheckURL.Method = http.MethodGet
	}
	if c.CheckURL.Scheme == "" {
		c.CheckURL.Scheme = c.URL.Scheme
	}
	if c.CheckURL.Hostname == "" {
		c.CheckURL.Hostname = c.URL.Hostname
	}
	c.CheckURL.IP = c.URL.IP
	if c.CheckURL.Port == 0 {
		c.CheckURL.Port = c.URL.Port
	}

	return nil
}

func (c config) getStaticWeight(upstream.Server) int {
	return c.StaticWeight
}

func getHTTPClient() *http.Client {
	return http.DefaultClient
}

func getStaticStatus(s upstream.Server) upstream.ServerStatus {
	return upstream.ServerStatusOnline
}

func handleRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	if log.Enabled(req.Context(), log.LevelTrace) {
		log.Trace("forward the http request",
			"requestid", upstream.GetRequestID(req.Context(), req),
			"method", req.Method, "host", req.Host, "url", req.URL.String(),
			"err", err)
	}
	return resp, err
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

var _ upstream.WeightedServer = &httpServer{}

type httpServer struct {
	state nets.RuntimeState
	conf  atomic.Value
}

func (s *httpServer) loadConf() config { return s.conf.Load().(config) }
func (s *httpServer) Update(info interface{}) (err error) {
	conf, err := newConfig(info.(Config))
	if err == nil {
		s.conf.Store(conf)
	}
	return
}

func (s *httpServer) ID() string                      { return s.loadConf().id }
func (s *httpServer) Type() string                    { return "http" }
func (s *httpServer) Info() interface{}               { return s.loadConf().Config }
func (s *httpServer) Weight() int                     { return s.loadConf().getWeight(s) }
func (s *httpServer) Status() upstream.ServerStatus   { return s.loadConf().getStatus(s) }
func (s *httpServer) RuntimeState() nets.RuntimeState { return s.state.Clone() }

func (s *httpServer) Check(ctx context.Context) (err error) {
	conf := s.loadConf()
	req, err := conf.CheckURL.Request(ctx, http.MethodGet)
	if err != nil {
		return err
	}

	var resp *http.Response
	if conf.HTTPClient == nil {
		resp, err = http.DefaultClient.Do(req)
	} else {
		resp, err = conf.HTTPClient.GetHTTPClient().Do(req)
	}

	if resp != nil {
		io.CopyBuffer(io.Discard, resp.Body, make([]byte, 256))
		resp.Body.Close()
	}

	if err == nil && resp.StatusCode >= 400 {
		err = fmt.Errorf("unexpected status code '%d'", resp.StatusCode)
	}

	return
}

func (s *httpServer) Serve(ctx context.Context, req interface{}) (err error) {
	var (
		r *http.Request
		w http.ResponseWriter
	)

	if c, ok := req.(*reqresp.Context); ok {
		w = c.ResponseWriter
		r = c.Request
	} else {
		c = reqresp.GetContextFromCtx(ctx)
		w = c.ResponseWriter
		r = c.Request
	}

	s.state.Inc()
	defer func() {
		s.state.Dec()
		if err == nil {
			s.state.IncSuccess()
		}
	}()

	conf := s.loadConf()

	r = r.Clone(ctx)
	r.RequestURI = ""      // Pretend to be a client request.
	r.URL.Host = conf.addr // Dial to the upstream backend server.
	r.URL.Scheme = conf.URL.Scheme

	if conf.URL.Method != "" {
		r.Method = conf.URL.Method
	}

	if conf.host != "" {
		r.Host = conf.host // Override the header "Host"
	}

	if conf.URL.Path != "" {
		r.URL.RawPath = ""
		r.URL.Path = conf.URL.Path
	}

	if len(conf.queries) > 0 {
		if r.URL.RawQuery == "" {
			r.URL.RawQuery = conf.rawQuery
		} else if values := r.URL.Query(); len(values) == 0 {
			r.URL.RawQuery = conf.rawQuery
		} else {
			for k, vs := range conf.queries {
				values[k] = vs
			}
			r.URL.RawQuery = values.Encode()
		}
	}

	if len(conf.headers) > 0 {
		if len(r.Header) == 0 {
			r.Header = conf.headers.Clone()
		} else {
			for k, vs := range conf.headers {
				r.Header[k] = vs
			}
		}
	}

	resp, err := conf.handleRequest(conf.getClient(), r)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err == nil {
		err = conf.handleResponse(w, resp)
	}
	return err
}
