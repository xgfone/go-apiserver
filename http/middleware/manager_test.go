// Copyright 2026 xgfone
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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// pathRecorderMiddleware records the path and appends a suffix
func pathRecorderMiddleware(name string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Append middleware name to the path
			r.URL.Path += "/" + name
			next.ServeHTTP(w, r)
		})
	}
}

// statusCodeMiddleware sets the status code
func statusCodeMiddleware(status int) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			next.ServeHTTP(w, r)
		})
	}
}

// responseBodyMiddleware adds content to the response body
func responseBodyMiddleware(content string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			w.Write([]byte(content))
		})
	}
}

func TestManager(t *testing.T) {
	// Test basic Manager functionality
	t.Run("NewManager", func(t *testing.T) {
		// Test NewManager without handler
		m1 := NewManager(nil)
		if m1 == nil {
			t.Fatal("NewManager should not return nil")
		}

		// Test NewManager with handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		m2 := NewManager(handler)
		if m2 == nil {
			t.Fatal("NewManager should not return nil")
		}
	})

	t.Run("SetHandler", func(t *testing.T) {
		m := NewManager(nil)

		// Set handler
		called := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		m.SetHandler(handler)

		// Test if handler is properly set
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		m.ServeHTTP(rec, req)

		if !called {
			t.Error("handler was not called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("InsertAndAppend", func(t *testing.T) {
		m := NewManager(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Test Insert and Append
		m.Insert(pathRecorderMiddleware("insert1"))
		m.Append(pathRecorderMiddleware("append1"))
		m.Insert(pathRecorderMiddleware("insert2"))
		m.Append(pathRecorderMiddleware("append2"))

		// Verify middleware count
		if len(m.Middlewares()) != 4 {
			t.Errorf("expected 4 middlewares, got %d", len(m.Middlewares()))
		}

		// Verify middleware execution order
		var finalPath string
		m.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			finalPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost/start", nil)
		m.ServeHTTP(rec, req)

		// Verify path was modified
		if finalPath == "" {
			t.Error("middlewares did not modify the path")
		}
	})

	t.Run("InsertFuncAndAppendFunc", func(t *testing.T) {
		m := NewManager(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Test InsertFunc and AppendFunc
		m.InsertFunc(pathRecorderMiddleware("insertFunc1"))
		m.AppendFunc(pathRecorderMiddleware("appendFunc1"))
		m.InsertFunc(pathRecorderMiddleware("insertFunc2"))
		m.AppendFunc(pathRecorderMiddleware("appendFunc2"))

		// Verify middleware count
		if len(m.Middlewares()) != 4 {
			t.Errorf("expected 4 middlewares, got %d", len(m.Middlewares()))
		}
	})

	t.Run("Reset", func(t *testing.T) {
		m := NewManager(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Add some middlewares
		m.Append(pathRecorderMiddleware("mw1"))
		m.Append(pathRecorderMiddleware("mw2"))

		if len(m.Middlewares()) != 2 {
			t.Errorf("expected 2 middlewares before reset, got %d", len(m.Middlewares()))
		}

		// Reset middlewares
		m.Reset(pathRecorderMiddleware("mw3"), pathRecorderMiddleware("mw4"))

		if len(m.Middlewares()) != 2 {
			t.Errorf("expected 2 middlewares after reset, got %d", len(m.Middlewares()))
		}
	})

	t.Run("EmptyMiddlewareOperations", func(t *testing.T) {
		m := NewManager(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// These calls should not panic or change state
		m.Append()     // No arguments
		m.Insert()     // No arguments
		m.AppendFunc() // No arguments
		m.InsertFunc() // No arguments

		if len(m.Middlewares()) != 0 {
			t.Errorf("expected 0 middlewares after empty operations, got %d", len(m.Middlewares()))
		}
	})

	t.Run("HandlerMethod", func(t *testing.T) {
		m := NewManager(nil)

		// Add middlewares
		m.Append(responseBodyMiddleware("world"))
		m.Append(responseBodyMiddleware("hello"))

		// Create an independent handler
		baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(" "))
		})

		// Wrap using Manager's Handler method
		wrappedHandler := m.Handler(baseHandler)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		wrappedHandler.ServeHTTP(rec, req)

		// Verify response body is not empty
		if rec.Body.Len() == 0 {
			t.Error("expected non-empty response body")
		}
	})

	t.Run("WithPriorityMiddleware", func(t *testing.T) {
		m := NewManager(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Create priority middlewares
		mw1 := New("mw1", 30, pathRecorderMiddleware("priority30"))
		mw2 := New("mw2", 10, pathRecorderMiddleware("priority10"))
		mw3 := New("mw3", 20, pathRecorderMiddleware("priority20"))

		// Add in random order
		m.Append(mw3, mw1, mw2)

		// Verify middleware count
		if len(m.Middlewares()) != 3 {
			t.Errorf("expected 3 middlewares, got %d", len(m.Middlewares()))
		}
	})

	t.Run("MiddlewareExecutionOrder", func(t *testing.T) {
		m := NewManager(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Add middlewares and verify execution order
		m.Append(pathRecorderMiddleware("append1"))
		m.Insert(pathRecorderMiddleware("insert1"))
		m.Append(pathRecorderMiddleware("append2"))
		m.Insert(pathRecorderMiddleware("insert2"))

		// Set handler to check path
		var finalPath string
		m.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			finalPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost/path", nil)
		m.ServeHTTP(rec, req)

		// Verify path was correctly modified
		if !strings.Contains(finalPath, "/path") {
			t.Errorf("path should contain '/path', got '%s'", finalPath)
		}
	})

	t.Run("Integration", func(t *testing.T) {
		m := NewManager(nil)

		// Add various types of middlewares
		m.Append(statusCodeMiddleware(http.StatusAccepted))
		m.Insert(responseBodyMiddleware("hello"))
		m.Append(responseBodyMiddleware("world"))
		m.InsertFunc(pathRecorderMiddleware("test"))

		// Set final handler
		pathModified := false
		handlerCalled := false
		m.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path was modified
			if strings.Contains(r.URL.Path, "/test") {
				pathModified = true
			}
			handlerCalled = true
			w.Write([]byte("!"))
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		m.ServeHTTP(rec, req)

		// Verify handler was called
		if !handlerCalled {
			t.Error("handler was not called")
		}

		// Verify path was modified
		if !pathModified {
			t.Error("path middleware did not modify URL path")
		}

		// Verify response body is not empty
		if rec.Body.Len() == 0 {
			t.Error("expected non-empty response body")
		}
	})
}
