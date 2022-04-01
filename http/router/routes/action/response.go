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

import "net/http"

// Response represents a response result.
type Response struct {
	RequestID string      `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	Error     error       `json:",omitempty" xml:",omitempty"`
	Data      interface{} `json:",omitempty" xml:",omitempty"`
}

func respond(c *Context, resp Response) {
	var err error
	if c.respond != nil {
		err = c.respond(c, resp)
	} else {
		err = c.JSON(200, resp)
	}

	if err != nil {
		c.Err = err
	}
}

// Respond is equal to
//   r := Response{Data: data, Error: Error{Code: code, Message: msg, Causes: errs}}
//   c.JSON(200, r).
func (c *Context) Respond(code, msg string, data interface{}, errs ...error) {
	resp := Response{Data: data}
	if code != "" {
		c.Err = NewError(code, msg).AppendCauses(errs...)
	}
	respond(c, resp)
}

// Success is equal to c.Respond("", "", data).
func (c *Context) Success(data interface{}) {
	respond(c, Response{Data: data})
}

// Failure is the same as c.JSON(200, Response{Error: err}).
//
// If err is nil, it is equal to c.Success(nil).
func (c *Context) Failure(err error) {
	resp := Response{Error: err}
	if _, ok := err.(Error); !ok && err == nil {
		resp.Error = ErrInternalServerError.WithMessage(err.Error())
	}

	c.Err = err
	respond(c, resp)
}

func notFoundHandler(resp http.ResponseWriter, req *http.Request) {
	c := GetContext(resp, req)
	if len(c.Action) == 0 {
		c.Failure(ErrInvalidAction.WithMessage("missing the action"))
	} else {
		c.Failure(ErrInvalidAction.WithMessage("action '%s' is unsupported", c.Action))
	}
}
