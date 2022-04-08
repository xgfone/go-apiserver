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
	"encoding/json"
	"io"
	"net/http"
)

// Response represents a response result.
type Response struct {
	RequestID string      `json:"RequestId,omitempty" yaml:"RequestId,omitempty"`
	Error     error       `json:"Error,omitempty" yaml:"Error,omitempty"`
	Data      interface{} `json:"Data,omitempty" yaml:"Data,omitempty"`
}

// Respond sends the response by the context as JSON.
func (r Response) Respond(c *Context) { c.Respond(r) }

// WithRequestID returns a new Response with the given request id.
func (r Response) WithRequestID(requestID string) Response {
	r.RequestID = requestID
	return r
}

// WithData returns a new Response with the given data.
func (r Response) WithData(data interface{}) Response {
	r.Data = data
	return r
}

// WithError returns a new Response with the given error.
func (r Response) WithError(err error) Response {
	r.Error = err
	return r
}

// Decode uses the decode function to decode the result to the response.
func (r *Response) Decode(decode func(interface{}) error) (err error) {
	return decode(r)
}

// DecodeJSON uses json decoder to decode from the reader into the response.
func (r *Response) DecodeJSON(reader io.Reader) (err error) {
	return json.NewDecoder(reader).Decode(r)
}

// DecodeJSONBytes uses json decoder to decode the []byte data into the response.
func (r *Response) DecodeJSONBytes(data []byte) (err error) {
	return json.Unmarshal(data, r)
}

// Respond is the same as c.JSON(200, response).
func (c *Context) Respond(response Response) {
	var err error
	if c.respond != nil {
		err = c.respond(c, response)
	} else {
		err = c.JSON(200, response)
	}

	if err != nil && c.Err == nil {
		c.Err = err
	}
}

// Success is equal to c.Respond(Response{Data: data}).
func (c *Context) Success(data interface{}) { c.Respond(Response{Data: data}) }

// Failure is the same as c.JSON(200, Response{Error: err}).
//
// If err is nil, it is equal to c.Success(nil).
func (c *Context) Failure(err error) {
	c.Err = err

	switch err.(type) {
	case nil, Error:
	default:
		err = ErrInternalServerError.WithError(err)
	}

	c.Respond(Response{Error: err})
}

func notFoundHandler(resp http.ResponseWriter, req *http.Request) {
	c := GetContext(resp, req)
	if len(c.Action) == 0 {
		c.Failure(ErrInvalidAction.WithMessage("missing the action"))
	} else {
		c.Failure(ErrInvalidAction.WithMessage("action '%s' is unsupported", c.Action))
	}
}
