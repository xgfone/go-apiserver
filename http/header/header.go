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

// Package header provides some header constants and operations.
package header

import (
	"net/http"

	"github.com/xgfone/go-toolkit/httpx"
)

var (
	mimeTextXML                = []string{httpx.MIMETextXML}
	mimeTextHTML               = []string{httpx.MIMETextHTML}
	mimeTextPlain              = []string{httpx.MIMETextPlain}
	mimeApplicationXML         = []string{httpx.MIMEApplicationXML}
	mimeApplicationJSON        = []string{httpx.MIMEApplicationJSON}
	mimeApplicationForm        = []string{httpx.MIMEApplicationForm}
	mimeApplicationMsgpack     = []string{httpx.MIMEApplicationMsgpack}
	mimeApplicationProtobuf    = []string{httpx.MIMEApplicationProtobuf}
	mimeApplicationOctetStream = []string{httpx.MIMEApplicationOctetStream}
	mimeMultipartForm          = []string{httpx.MIMEMultipartForm}

	mimeTextXMLCharsetUTF8         = []string{httpx.MIMETextXMLCharsetUTF8}
	mimeTextHTMLCharsetUTF8        = []string{httpx.MIMETextHTMLCharsetUTF8}
	mimeTextPlainCharsetUTF8       = []string{httpx.MIMETextPlainCharsetUTF8}
	mimeApplicationXMLCharsetUTF8  = []string{httpx.MIMEApplicationXMLCharsetUTF8}
	mimeApplicationJSONCharsetUTF8 = []string{httpx.MIMEApplicationJSONCharsetUTF8}
)

// SetContentType sets the header "Content-Type" to ct.
//
// If ct is "", do nothing.
func SetContentType(header http.Header, ct string) {
	switch ct {
	case "":

	case httpx.MIMETextXML:
		header[httpx.HeaderContentType] = mimeTextXML

	case httpx.MIMETextHTML:
		header[httpx.HeaderContentType] = mimeTextHTML

	case httpx.MIMETextPlain:
		header[httpx.HeaderContentType] = mimeTextPlain

	case httpx.MIMEApplicationXML:
		header[httpx.HeaderContentType] = mimeApplicationXML

	case httpx.MIMEApplicationJSON:
		header[httpx.HeaderContentType] = mimeApplicationJSON

	case httpx.MIMEApplicationForm:
		header[httpx.HeaderContentType] = mimeApplicationForm

	case httpx.MIMEApplicationMsgpack:
		header[httpx.HeaderContentType] = mimeApplicationMsgpack

	case httpx.MIMEApplicationProtobuf:
		header[httpx.HeaderContentType] = mimeApplicationProtobuf

	case httpx.MIMEApplicationOctetStream:
		header[httpx.HeaderContentType] = mimeApplicationOctetStream

	case httpx.MIMEMultipartForm:
		header[httpx.HeaderContentType] = mimeMultipartForm

	case httpx.MIMETextXMLCharsetUTF8:
		header[httpx.HeaderContentType] = mimeTextXMLCharsetUTF8

	case httpx.MIMETextHTMLCharsetUTF8:
		header[httpx.HeaderContentType] = mimeTextHTMLCharsetUTF8

	case httpx.MIMETextPlainCharsetUTF8:
		header[httpx.HeaderContentType] = mimeTextPlainCharsetUTF8

	case httpx.MIMEApplicationXMLCharsetUTF8:
		header[httpx.HeaderContentType] = mimeApplicationXMLCharsetUTF8

	case httpx.MIMEApplicationJSONCharsetUTF8:
		header[httpx.HeaderContentType] = mimeApplicationJSONCharsetUTF8

	default:
		header.Set(httpx.HeaderContentType, ct)
	}
}
