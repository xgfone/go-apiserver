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

package render

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func ExampleMuxRenderer() {
	mr := NewMuxRenderer()

	// JSON
	mr.Add("json", RendererFunc(func(w http.ResponseWriter, code int, name string, data interface{}) error {
		buf := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buf).Encode(data); err != nil {
			return err
		}

		w.WriteHeader(code)
		w.Write(buf.Bytes())
		return nil
	}))

	// XML
	mr.Add("xml", RendererFunc(func(w http.ResponseWriter, code int, name string, data interface{}) error {
		buf := bytes.NewBuffer(nil)
		if err := xml.NewEncoder(buf).Encode(data); err != nil {
			return err
		}

		w.WriteHeader(code)
		w.Write(buf.Bytes())
		return nil
	}))

	type Response struct {
		Result string
	}
	response := Response{Result: "OK"}

	recjson := httptest.NewRecorder()
	mr.Render(recjson, 200, "json", response)
	fmt.Println(recjson.Body.String())

	recxml := httptest.NewRecorder()
	mr.Render(recxml, 200, "xml", response)
	fmt.Println(recxml.Body.String())

	// Output:
	// {"Result":"OK"}
	//
	// <Response><Result>OK</Result></Response>
}
