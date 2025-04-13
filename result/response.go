// Copyright 2022~2023 xgfone
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

package result

import (
	"bytes"
	"fmt"
	"io"

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-toolkit/jsonx"
)

// Respond is the public function to send the response by responder.
var Respond func(responder any, response Response) = DefaultRespond

// Success is a convenient function, which is equal to
//
//	Respond(responder, Ok(data))
func Success(responder, data any) {
	Respond(responder, Ok(data))
}

// Failure is a convenient function, which is equal to
//
//	Respond(responder, Err(err))
func Failure(responder any, err error) {
	Respond(responder, Err(err))
}

// Response represents a response result.
type Response struct {
	Error error `json:",omitempty"`
	Data  any   `json:",omitempty"`
}

// NewResponse returns a new response.
func NewResponse(data any, err error) Response {
	return Response{Data: data, Error: err}
}

// Ok is equal to NewResponse(data, nil).
func Ok(data any) Response { return Response{Data: data} }

// Err is equal to NewResponse(nil, err).
func Err(err error) Response { return Response{Error: err} }

// IsZero reports whether the response is ZERO.
func (r Response) IsZero() bool {
	return r == (Response{})
}

// WithData returns a new Response with the given data.
func (r Response) WithData(data any) Response {
	r.Data = data
	return r
}

// WithError returns a new Response with the given error.
func (r Response) WithError(err error) Response {
	r.Error = err
	return r
}

// Decode uses the decode function to decode the result to the response.
func (r *Response) Decode(decode func(any) error) error {
	return decode(r)
}

// DecodeJSON uses json decoder to decode from the reader into the response.
func (r *Response) DecodeJSON(reader io.Reader) error {
	return jsonx.UnmarshalReader(r, reader)
}

// DecodeJSONBytes uses json decoder to decode the []byte data into the response.
func (r *Response) DecodeJSONBytes(data []byte) error {
	return jsonx.UnmarshalReader(r, bytes.NewReader(data))
}

// Respond sends the response by the responder,
// which will forward the calling to Respond.
func (r Response) Respond(responder any) {
	Respond(responder, r)
}

// StatusCode inspects and returns the status code by the error.
func (r Response) StatusCode() int {
	if r.Error == nil {
		return 200
	}

	if v, ok := r.Error.(interface{ StatusCode() int }); ok {
		return v.StatusCode()
	}

	return 500
}

// Responder is the responder interface.
type Responder interface {
	Respond(Response)
}

// DefaultRespond is the default implemention to send the response by responder,
// which will try to assert responder to the types as follows, and call it:
//
//	Responder
//	handler.JSONResponder
//
// For other types, it will panic.
func DefaultRespond(responder any, response Response) {
	switch resp := responder.(type) {
	case Responder:
		resp.Respond(response)

	case handler.JSONResponder:
		if response.IsZero() {
			resp.JSON(response.StatusCode(), nil)
		} else {
			resp.JSON(response.StatusCode(), response)
		}

	default:
		panic(fmt.Errorf("result.DefaultRespond: unknown responder type %T", responder))
	}
}
