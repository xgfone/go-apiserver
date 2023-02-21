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

// Package balancer implements some balancer forwarding policies.
package balancer

import (
	"context"

	"github.com/xgfone/go-apiserver/upstream"
)

// Balancer is used to forward the request to one of the backend servers.
type Balancer interface {
	Forward(ctx context.Context, req interface{}, sd upstream.ServerDiscovery) error
	Policy() string
}

// ForwardFunc is the function to forward the request to one of the servers.
type ForwardFunc func(context.Context, interface{}, upstream.ServerDiscovery) error

// NewBalancer returns a new balancer with the policy and the forward function.
func NewBalancer(policy string, forward ForwardFunc) Balancer {
	return balancer{forward: forward, policy: policy}
}

type balancer struct {
	forward ForwardFunc
	policy  string
}

func (b balancer) Policy() string { return b.policy }
func (b balancer) Forward(c context.Context, r interface{}, sd upstream.ServerDiscovery) error {
	return b.forward(c, r, sd)
}
