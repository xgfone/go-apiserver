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

package header

// MIME types
const (
	MIMETextXML                = "text/xml"
	MIMETextHTML               = "text/html"
	MIMETextPlain              = "text/plain"
	MIMEApplicationXML         = "application/xml"
	MIMEApplicationJSON        = "application/json"
	MIMEApplicationProtobuf    = "application/protobuf"
	MIMEApplicationMsgpack     = "application/msgpack"
	MIMEApplicationOctetStream = "application/octet-stream"
	MIMEApplicationForm        = "application/x-www-form-urlencoded"
	MIMEMultipartForm          = "multipart/form-data"

	MIMETextXMLCharsetUTF8         = MIMETextXML + "; charset=UTF-8"
	MIMETextHTMLCharsetUTF8        = MIMETextHTML + "; charset=UTF-8"
	MIMETextPlainCharsetUTF8       = MIMETextPlain + "; charset=UTF-8"
	MIMEApplicationXMLCharsetUTF8  = MIMEApplicationXML + "; charset=UTF-8"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; charset=UTF-8"
)

// Headers
const (
	HeaderAllow               = "Allow"
	HeaderAccept              = "Accept"
	HeaderAcceptedLanguage    = "Accept-Language"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAuthorization       = "Authorization"
	HeaderConnection          = "Connection"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCacheControl        = "Cache-Control"
	HeaderIfNoneMatch         = "If-None-Match"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderExpires             = "Expires"
	HeaderEtag                = "Etag"
	HeaderVary                = "Vary"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderUpgrade             = "Upgrade"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"
	HeaderReferer             = "Referer"
	HeaderLocation            = "Location"
	HeaderUserAgent           = "User-Agent"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSSL       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-Ip"
	HeaderXServerID           = "X-Server-Id"
	HeaderXRequestID          = "X-Request-Id"
	HeaderXRequestedWith      = "X-Requested-With"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-Xss-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderXCSRFToken              = "X-Csrf-Token"
)
