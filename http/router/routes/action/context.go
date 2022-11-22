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

package action

import (
	"net/http"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/result"
)

// GetContext returns the Context from the http request.
func GetContext(w http.ResponseWriter, r *http.Request) *Context {
	c, _ := reqresp.GetContext(w, r).Reg3.(*Context)
	return c
}

var _ result.Responder = new(Context)

// Context is the request context.
type Context struct {
	// Notice: It uses Reg3 to store the action context.
	*reqresp.Context

	// If action is empty, it represents that there is no action in the request.
	Action  string
	handler http.Handler
	respond func(*Context, result.Response) error
}

// Reset resets the context.
func (c *Context) Reset() { *c = Context{} }

// Respond is the same as c.JSON(200, response).
func (c *Context) Respond(response result.Response) {
	if response.Error != nil {
		c.UpdateError(response.Error)
	}

	var err error
	if c.respond != nil {
		err = c.respond(c, response)
	} else {
		err = c.JSON(200, response)
	}
	c.UpdateError(err)
}

// Success is equal to c.Respond(Response{Data: data}).
func (c *Context) Success(data interface{}) {
	c.Respond(result.Response{Data: data})
}

// Failure is the same as c.JSON(200, Response{Error: err}).
//
// If err is nil, it is equal to c.Success(nil).
func (c *Context) Failure(err error) {
	switch err.(type) {
	case nil, result.Error:
	default:
		err = result.ErrInternalServerError.WithError(err)
	}
	c.Respond(result.Response{Error: err})
}

func notFoundHandler(resp http.ResponseWriter, req *http.Request) {
	c := GetContext(resp, req)
	if len(c.Action) == 0 {
		c.Failure(result.ErrInvalidAction.WithMessage("missing the action"))
	} else {
		c.Failure(result.ErrInvalidAction.WithMessage("action '%s' is unsupported", c.Action))
	}
}
