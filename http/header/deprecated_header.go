// Copyright 2021~2025 xgfone
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

package header

import (
	"net/http"

	"github.com/xgfone/go-toolkit/httpx"
)

// ContentType is the alias of the function MediaType.
//
// DEPRECATED: Use httpx.ContentType instead.
func ContentType(header http.Header) string {
	return httpx.ContentType(header)
}

// MediaType returns the MIME media type portion of the header "Content-Type".
//
// DEPRECATED: Use httpx.ContentType instead.
func MediaType(header http.Header) (mime string) {
	return httpx.ContentType(header)
}

// Charset returns the charset of the request content.
//
// Return "" if there is no charset.
//
// DEPRECATED: Use httpx.Charset instead.
func Charset(header http.Header) string {
	return httpx.Charset(header)
}

// IsWebSocket reports whether the request is websocket.
//
// DEPRECATED: Use httpx.IsWebSocket instead.
func IsWebSocket(req *http.Request) bool {
	return httpx.IsWebSocket(req)
}

// Accept returns the accepted Content-Type list from the request header
// "Accept", which are sorted by the q-factor weight from high to low.
//
// If there is no the request header "Accept", return nil.
//
// Notice:
//  1. If the value is "*/*", it will be amended as "".
//  2. If the value is "<MIME_type>/*", it will be amended as "<MIME_type>/".
//     So it can be used to match the prefix.
//
// DEPRECATED: Use httpx.Accept instead.
func Accept(header http.Header) []string {
	return httpx.Accept(header)
}

// Scheme returns the HTTP protocol scheme, `http` or `https`.
//
// DEPRECATED.
func Scheme(header http.Header) (scheme string) {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if header.Get(HeaderXForwardedSSL) == "on" {
		return "https"
	} else if scheme = header.Get(HeaderXForwardedProto); scheme != "" {
		return
	} else if scheme = header.Get(HeaderXForwardedProtocol); scheme != "" {
		return
	} else if scheme = header.Get(HeaderXUrlScheme); scheme != "" {
		return
	}

	return "http"
}
