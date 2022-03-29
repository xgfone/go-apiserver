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

package service

import (
	"context"
	"net/http"
)

// NewURLChecker returns a new url checker that checks whether to access the url
// with the method GET returns the status code 2xx.
func NewURLChecker(rawURL string) (Checker, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	return urlChecker{url: rawURL, req: req}, nil
}

// MustURLChecker is the same as NewURLChecker, but panics if there is an error.
func MustURLChecker(rawURL string) Checker {
	checker, err := NewURLChecker(rawURL)
	if err != nil {
		panic(err)
	}
	return checker
}

type urlChecker struct {
	req *http.Request
	url string
}

func (c urlChecker) Name() string { return "url:" + c.url }
func (c urlChecker) Check(ctx context.Context) (ok bool, err error) {
	resp, err := http.DefaultClient.Do(c.req.WithContext(ctx))
	if resp != nil {
		resp.Body.Close()
	}

	if err == nil {
		ok = resp.StatusCode >= 200 && resp.StatusCode < 300
	}

	return
}
