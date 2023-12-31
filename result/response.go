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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/xgfone/go-apiserver/http/handler"
)

// Respond is used to send the response by responder,
// such as http.ResponseWriter.
var Respond func(responder any, response Response)

// Response represents a response result.
type Response struct {
	Error error       `json:"error,omitempty" yaml:"error,omitempty" xml:"error,omitempty"`
	Data  interface{} `json:"data,omitempty" yaml:"data,omitempty" xml:"data,omitempty"`
}

// NewResponse returns a new response.
func NewResponse(data interface{}, err error) Response {
	return Response{Data: data, Error: err}
}

// Ok is equal to NewResponse(data, nil).
func Ok(data interface{}) Response { return Response{Data: data} }

// Err is equal to NewResponse(nil, err).
func Err(err error) Response { return Response{Error: err} }

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
func (r *Response) Decode(decode func(interface{}) error) error {
	return decode(r)
}

// DecodeJSON uses json decoder to decode from the reader into the response.
func (r *Response) DecodeJSON(reader io.Reader) error {
	return json.NewDecoder(reader).Decode(r)
}

// DecodeJSONBytes uses json decoder to decode the []byte data into the response.
func (r *Response) DecodeJSONBytes(data []byte) error {
	return json.Unmarshal(data, r)
}

// Respond sends the response by the responder.
//
// If Respond is set, forward it with responder and response to handle.
// If not set, it tries to assert responder to one of types as follow:
//
//	interface{ Respond(Response) }
//	interface{ JSON(code int, value interface{}) }
//	http.ResponseWriter
func (r Response) Respond(responder any) {
	if Respond != nil {
		Respond(responder, r)
		return
	}

	if r.Data == nil && r.Error == nil {
		return
	}

	switch resp := responder.(type) {
	case interface{ Respond(Response) }:
		resp.Respond(r)

	case interface{ JSON(int, interface{}) }:
		resp.JSON(getStatusCode(r.Error), r)

	case http.ResponseWriter:
		sendjson(resp, r)

	default:
		panic(fmt.Errorf("Response.Respond: unknown responder type %T", responder))
	}
}

func getStatusCode(err error) int {
	if v, ok := err.(interface{ StatusCode() int }); ok {
		return v.StatusCode()
	}
	return 200
}

func sendjson(w http.ResponseWriter, v Response) {
	err := handler.JSON(w, getStatusCode(v.Error), v)
	if err != nil {
		slog.Error("fail to encode and send response to client", "err", err)
	}
}
