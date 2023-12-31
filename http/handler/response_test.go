// Copyright 2024 xgfone
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

package handler

import (
	"encoding/xml"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	err := JSON(rec, 400, map[string]string{"a": "b"})
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != 400 {
		t.Errorf("expect status code %d, but got %d", 400, rec.Code)
	}

	expectbody := `{"a":"b"}` + "\n"
	if body := rec.Body.String(); body != expectbody {
		t.Errorf("expect response body '%s', but got '%s'", expectbody, body)
	}
}

func TestXML(t *testing.T) {
	var req struct {
		XMLName xml.Name `xml:"outer"`
		A       string   `xml:"a"`
	}
	req.A = "b"

	rec := httptest.NewRecorder()
	err := XML(rec, 400, req)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != 400 {
		t.Errorf("expect status code %d, but got %d", 400, rec.Code)
	}

	expectbody := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<outer><a>b</a></outer>`
	if body := rec.Body.String(); body != expectbody {
		t.Errorf("expect response body '%s', but got '%s'", expectbody, body)
	}
}
