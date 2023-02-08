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

package logger

import "context"

// Define the global logger option config.
var (
	LogReqBodyLen  func(context.Context) int
	LogRespBodyLen func(context.Context) int

	LogReqHeaders  func(context.Context) bool
	LogRespHeaders func(context.Context) bool
)

// Config is used to configure the logger middleware.
type Config struct {
	// For Common
	LogReqBodyLen  func(context.Context) int
	LogRespBodyLen func(context.Context) int
	LogReqHeaders  func(context.Context) bool
	LogRespHeaders func(context.Context) bool

	// For HTTP
	LogHTTPReqBodyLen  func(context.Context) int
	LogHTTPRespBodyLen func(context.Context) int
	LogHTTPReqHeaders  func(context.Context) bool
	LogHTTPRespHeaders func(context.Context) bool
}

// GetLogReqBodyLen returns the maximum length of the request body to be logged.
// If 0 or a negative, not log the request body.
//
// Notice: it logs the request body only if the length of the request body
// is not greater than the returned length.
//
// If the field LogReqBody is nil, call the global function LogReqBody instead.
// If it's also nil, return 0 instead.
func (c Config) GetLogReqBodyLen(ctx context.Context) int {
	if c.LogReqBodyLen != nil {
		return c.LogReqBodyLen(ctx)
	} else if LogReqBodyLen != nil {
		return LogReqBodyLen(ctx)
	}
	return 0
}

// GetLogRespBodyLen returns the maximum length of the response body to be logged.
// If 0 or a negative, not log the response body.
//
// Notice: it logs the response body only if the length of the response body
// is not greater than the returned length.
//
// If the field LogRespBodyLen is nil, call the global function LogRespBodyLen
// instead. If it's also nil, return 0 instead.
func (c Config) GetLogRespBodyLen(ctx context.Context) int {
	if c.LogRespBodyLen != nil {
		return c.LogRespBodyLen(ctx)
	} else if LogRespBodyLen != nil {
		return LogRespBodyLen(ctx)
	}
	return 0
}

// GetLogReqHeaders reports whether to log the request headers.
//
// If the field LogReqHeaders is nil, call the global function LogReqHeaders
// instead. If it's also nil, return false instead.
func (c Config) GetLogReqHeaders(ctx context.Context) bool {
	if c.LogReqHeaders != nil {
		return c.LogReqHeaders(ctx)
	} else if LogReqHeaders != nil {
		return LogReqHeaders(ctx)
	}
	return false
}

// GetLogRespHeaders reports whether to log the response headers.
//
// If the field LogRespHeaders is nil, call the global function LogRespHeaders
// instead. If it's also nil, return 0 instead.
func (c Config) GetLogRespHeaders(ctx context.Context) bool {
	if c.LogRespHeaders != nil {
		return c.LogRespHeaders(ctx)
	} else if LogRespHeaders != nil {
		return LogRespHeaders(ctx)
	}
	return false
}

// Option is the logger option to update the logger config.
type Option func(*Config)

// SetLogReqBodyLen returns a logger option to set the field LogReqBodyLen.
func SetLogReqBodyLen(f func(context.Context) int) Option {
	return func(lc *Config) { lc.LogReqBodyLen = f }
}

// SetLogRespBodyLen returns a logger option to set the field LogRespBodyLen.
func SetLogRespBodyLen(f func(context.Context) int) Option {
	return func(lc *Config) { lc.LogRespBodyLen = f }
}

// SetLogRespHeaders returns a logger option to set the field LogRespHeaders.
func SetLogRespHeaders(f func(context.Context) bool) Option {
	return func(lc *Config) { lc.LogRespHeaders = f }
}

// SetLogReqHeaders returns a logger option to set the field LogReqHeaders.
func SetLogReqHeaders(f func(context.Context) bool) Option {
	return func(lc *Config) { lc.LogReqHeaders = f }
}
