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

package client

import (
	"io"
	"net/http"

	"github.com/xgfone/go-apiserver/internal/atomic"
)

// Getter is used to get the http client.
type Getter interface {
	GetHTTPClient() *http.Client
}

// Setter is used to set the http client.
type Setter interface {
	SetHTTPClient(*http.Client)
}

// Client is used to maintain the http.Client thread-safely.
type Client struct {
	httpClient atomic.Value
}

// NewClient returns a new thread-safe http client.
func NewClient(client *http.Client) *Client {
	if client == nil {
		panic("the http client is nil")
	}

	c := new(Client)
	c.SetHTTPClient(client)
	return c
}

// GetHTTPClient implements the interface Getter to get the http.Client.
func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient.Load().(*http.Client)
}

// SetHTTPClient implements the interface Setter to set the http.Client.
func (c *Client) SetHTTPClient(client *http.Client) {
	if client == nil {
		panic("the http client is nil")
	}
	c.httpClient.Store(client)
}

// SwapHTTPClient swaps the old http.Client out with the new.
func (c *Client) SwapHTTPClient(new *http.Client) (old *http.Client) {
	if new == nil {
		panic("the http client is nil")
	}
	return c.httpClient.Swap(new).(*http.Client)
}

// Do is a convenient function to send the http request.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.GetHTTPClient().Do(req)
}

// Get is a convenient function to send the http GET request.
func (c *Client) Get(url string) (*http.Response, error) {
	return c.GetHTTPClient().Get(url)
}

// Post is a convenient function to send the http POST request.
func (c *Client) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return c.GetHTTPClient().Post(url, contentType, body)
}
