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

// Package header provides some header constants and operations.
package header

import (
	"net/http"
	"strings"
)

var (
	mimeTextXML                = []string{MIMETextXML}
	mimeTextHTML               = []string{MIMETextHTML}
	mimeTextPlain              = []string{MIMETextPlain}
	mimeApplicationXML         = []string{MIMEApplicationXML}
	mimeApplicationJSON        = []string{MIMEApplicationJSON}
	mimeApplicationForm        = []string{MIMEApplicationForm}
	mimeApplicationMsgpack     = []string{MIMEApplicationMsgpack}
	mimeApplicationProtobuf    = []string{MIMEApplicationProtobuf}
	mimeApplicationOctetStream = []string{MIMEApplicationOctetStream}
	mimeMultipartForm          = []string{MIMEMultipartForm}

	mimeTextXMLCharsetUTF8         = []string{MIMETextXMLCharsetUTF8}
	mimeTextHTMLCharsetUTF8        = []string{MIMETextHTMLCharsetUTF8}
	mimeTextPlainCharsetUTF8       = []string{MIMETextPlainCharsetUTF8}
	mimeApplicationXMLCharsetUTF8  = []string{MIMEApplicationXMLCharsetUTF8}
	mimeApplicationJSONCharsetUTF8 = []string{MIMEApplicationJSONCharsetUTF8}
)

// SetContentType sets the header "Content-Type" to ct.
//
// If ct is "", do nothing.
func SetContentType(header http.Header, ct string) {
	switch ct {
	case "":

	case MIMETextXML:
		header[HeaderContentType] = mimeTextXML

	case MIMETextHTML:
		header[HeaderContentType] = mimeTextHTML

	case MIMETextPlain:
		header[HeaderContentType] = mimeTextPlain

	case MIMEApplicationXML:
		header[HeaderContentType] = mimeApplicationXML

	case MIMEApplicationJSON:
		header[HeaderContentType] = mimeApplicationJSON

	case MIMEApplicationForm:
		header[HeaderContentType] = mimeApplicationForm

	case MIMEApplicationMsgpack:
		header[HeaderContentType] = mimeApplicationMsgpack

	case MIMEApplicationProtobuf:
		header[HeaderContentType] = mimeApplicationProtobuf

	case MIMEApplicationOctetStream:
		header[HeaderContentType] = mimeApplicationOctetStream

	case MIMEMultipartForm:
		header[HeaderContentType] = mimeMultipartForm

	case MIMETextXMLCharsetUTF8:
		header[HeaderContentType] = mimeTextXMLCharsetUTF8

	case MIMETextHTMLCharsetUTF8:
		header[HeaderContentType] = mimeTextHTMLCharsetUTF8

	case MIMETextPlainCharsetUTF8:
		header[HeaderContentType] = mimeTextPlainCharsetUTF8

	case MIMEApplicationXMLCharsetUTF8:
		header[HeaderContentType] = mimeApplicationXMLCharsetUTF8

	case MIMEApplicationJSONCharsetUTF8:
		header[HeaderContentType] = mimeApplicationJSONCharsetUTF8

	default:
		header.Set(HeaderContentType, ct)
	}
}

// ContentType is the alias of the function MediaType.
func ContentType(header http.Header) string { return MediaType(header) }

// MediaType returns the MIME media type portion of the header "Content-Type".
func MediaType(header http.Header) (mime string) {
	mime = header.Get(HeaderContentType)
	if index := strings.IndexAny(mime, ";"); index > 0 {
		mime = strings.TrimSpace(mime[:index])
	}
	return
}

// Charset returns the charset of the request content.
//
// Return "" if there is no charset.
func Charset(header http.Header) string {
	ct := header.Get(HeaderContentType)
	for loop := len(ct) > 0; loop; {
		index := strings.IndexByte(ct, ';')
		if loop = index > 0; loop {
			ct = ct[index+1:]
		}

		if index = strings.IndexByte(ct, '='); index > 0 {
			if strings.TrimSpace(ct[:index]) == "charset" {
				return strings.TrimSpace(ct[index+1:])
			}
		}
	}
	return ""
}

// IsWebSocket reports whether the request is websocket.
func IsWebSocket(req *http.Request) bool {
	return req.Method == http.MethodGet &&
		req.Header.Get(HeaderConnection) == HeaderUpgrade &&
		req.Header.Get(HeaderUpgrade) == "websocket"
}
