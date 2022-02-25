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

package html

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func filter(file string) bool { return !strings.HasSuffix(file, ".html") }

func TestDirLoader(t *testing.T) {
	loader := NewDirLoaderWithFilter(filter, ".")

	files, err := loader.LoadAll()
	if err != nil {
		t.Error(err)
	} else {
		for _, file := range files {
			switch name := file.Name(); name {
			case "template_test.html", "templates/div.html":
			default:
				t.Errorf("unexpected template file name '%s'", name)
			}
		}
	}
}

func TestTemplate(t *testing.T) {
	tmpl := NewTemplate(NewDirLoaderWithFilter(filter, "."))
	tmpl.Debug(true).Delims("{{", "}}")
	if err := tmpl.Reload(); err != nil {
		t.Errorf("fail to reload the template files: %s", err)
		return
	}

	data := map[string]string{"Data": "test"}
	html := `
<html>
    <body>
    <div>test</div>

    </body>
</html>
`

	rec := httptest.NewRecorder()
	if err := tmpl.Render(rec, 200, "template_test.html", data); err != nil {
		t.Errorf("fail to render the template: %s", err)
	} else if body := rec.Body.String(); body != html {
		t.Errorf("expect '%s', but got '%s'", html, body)
	}
}
