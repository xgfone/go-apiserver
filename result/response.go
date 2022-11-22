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

package result

import (
	"encoding/json"
	"io"
)

// Responder is used to send the result to the peer.
type Responder interface {
	Respond(Response)
}

// Response represents a response result.
type Response struct {
	RequestID string      `json:"RequestId,omitempty" yaml:"RequestId,omitempty" xml:"RequestId,omitempty"`
	Error     error       `json:",omitempty" yaml:",omitempty" xml:",omitempty"`
	Data      interface{} `json:",omitempty" yaml:",omitempty" xml:",omitempty"`
}

// NewResponse returns a new response.
func NewResponse(data interface{}, err error) Response {
	return Response{Data: data, Error: err}
}

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

// Respond sends the response by the context as JSON.
func (r Response) Respond(responder Responder) { responder.Respond(r) }

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
