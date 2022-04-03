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

// Package balancer implements some balancer forwarding policies.
package balancer

import (
	"net/http"

	"github.com/xgfone/go-apiserver/http/upstream"
)

// Balancer is used to forward the request to one of the backend servers.
type Balancer interface {
	Forward(http.ResponseWriter, *http.Request, upstream.Servers) error
	Policy() string
}

// ForwardFunc is the function to forward the request to one of the servers.
type ForwardFunc func(http.ResponseWriter, *http.Request, upstream.Servers) error

// NewBalancer returns a new balancer with the policy and the forward function.
func NewBalancer(policy string, forward ForwardFunc) Balancer {
	return balancer{forward: forward, policy: policy}
}

type balancer struct {
	forward ForwardFunc
	policy  string
}

func (f balancer) Policy() string { return f.policy }
func (f balancer) Forward(w http.ResponseWriter, r *http.Request, ss upstream.Servers) error {
	return f.forward(w, r, ss)
}
