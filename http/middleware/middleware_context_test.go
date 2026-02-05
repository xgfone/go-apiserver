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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

func TestOrMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		handlers    []ContextHandler
		expectError bool
		expectNext  bool
	}{
		{
			name:        "no handlers",
			handlers:    []ContextHandler{},
			expectError: false,
			expectNext:  true, // Empty handlers returns next directly
		},
		{
			name: "single handler success",
			handlers: []ContextHandler{
				func(c *reqresp.Context) error {
					c.Data = map[string]any{"handler": 1}
					return nil
				},
			},
			expectError: false,
			expectNext:  true, // Success calls next
		},
		{
			name: "single handler error",
			handlers: []ContextHandler{
				func(c *reqresp.Context) error {
					return errors.New("handler error")
				},
			},
			expectError: true,
			expectNext:  false, // All handlers fail, no next
		},
		{
			name: "multiple handlers first success",
			handlers: []ContextHandler{
				func(c *reqresp.Context) error {
					return nil // Success, should stop
				},
				func(c *reqresp.Context) error {
					t.Error("Second handler should not be called")
					return nil
				},
			},
			expectError: false,
			expectNext:  true, // First success calls next
		},
		{
			name: "multiple handlers second success",
			handlers: []ContextHandler{
				func(c *reqresp.Context) error {
					return errors.New("first error")
				},
				func(c *reqresp.Context) error {
					return nil // Success, should stop
				},
			},
			expectError: false,
			expectNext:  true, // Second success calls next
		},
		{
			name: "multiple handlers all fail",
			handlers: []ContextHandler{
				func(c *reqresp.Context) error {
					return errors.New("first error")
				},
				func(c *reqresp.Context) error {
					return errors.New("second error")
				},
			},
			expectError: true,
			expectNext:  false, // All handlers fail, no next
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Or("test-or", 10, tt.handlers...)

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			})

			handler := middleware.Handler(next)

			req := httptest.NewRequest("GET", "/test", nil)
			ctx := &reqresp.Context{
				Request: req,
			}
			req = req.WithContext(reqresp.SetContext(req.Context(), ctx))

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			// Check if next was called
			if nextCalled != tt.expectNext {
				t.Errorf("next called = %v, want %v", nextCalled, tt.expectNext)
			}

			// Check if error was set
			if tt.expectError && ctx.Err == nil {
				t.Error("expected error in context, got nil")
			} else if !tt.expectError && ctx.Err != nil {
				t.Errorf("unexpected error in context: %v", ctx.Err)
			}
		})
	}
}

func TestAndMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		handlers    []ContextHandler
		expectError bool
		expectNext  bool
	}{
		{
			name:        "no handlers",
			handlers:    []ContextHandler{},
			expectError: false,
			expectNext:  true, // Empty handlers returns next directly
		},
		{
			name: "single handler success",
			handlers: []ContextHandler{
				func(c *reqresp.Context) error {
					c.Data = map[string]any{"handler": 1}
					return nil
				},
			},
			expectError: false,
			expectNext:  true, // Single handler success means all handlers succeed, calls next
		},
		{
			name: "single handler error",
			handlers: []ContextHandler{
				func(c *reqresp.Context) error {
					return errors.New("handler error")
				},
			},
			expectError: true,
			expectNext:  false, // Error returns immediately, no next
		},
		{
			name: "multiple handlers all success",
			handlers: []ContextHandler{
				func(c *reqresp.Context) error {
					return nil // First success
				},
				func(c *reqresp.Context) error {
					return nil // Second success - both succeed, should call next
				},
			},
			expectError: false,
			expectNext:  true, // All handlers succeed, calls next
		},
		{
			name: "multiple handlers first fails",
			handlers: []ContextHandler{
				func(c *reqresp.Context) error {
					return errors.New("first error")
				},
				func(c *reqresp.Context) error {
					// This should not be called because first handler fails
					return nil
				},
			},
			expectError: true,
			expectNext:  false, // First error stops, no next
		},
		{
			name: "multiple handlers second fails",
			handlers: []ContextHandler{
				func(c *reqresp.Context) error {
					return nil // First success
				},
				func(c *reqresp.Context) error {
					return errors.New("second error") // Second fails
				},
			},
			expectError: true,
			expectNext:  false, // Second error stops, no next
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of handlers to avoid test interference
			handlers := make([]ContextHandler, len(tt.handlers))
			copy(handlers, tt.handlers)

			middleware := And("test-and", 10, handlers...)

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			})

			handler := middleware.Handler(next)

			req := httptest.NewRequest("GET", "/test", nil)
			ctx := &reqresp.Context{
				Request: req,
			}
			req = req.WithContext(reqresp.SetContext(req.Context(), ctx))

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			// Check if next was called
			if nextCalled != tt.expectNext {
				t.Errorf("next called = %v, want %v", nextCalled, tt.expectNext)
			}

			// Check if error was set
			if tt.expectError && ctx.Err == nil {
				t.Error("expected error in context, got nil")
			} else if !tt.expectError && ctx.Err != nil {
				t.Errorf("unexpected error in context: %v", ctx.Err)
			}
		})
	}
}

func TestMiddlewareContextMissing(t *testing.T) {
	middleware := Or("test-missing-context", 10,
		func(c *reqresp.Context) error {
			return nil
		},
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called when context is missing")
	})

	handler := middleware.Handler(next)

	req := httptest.NewRequest("GET", "/test", nil)
	// Don't add context to request

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	expectedBody := "missing reqresp.Context"
	if body := rec.Body.String(); body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}

func TestMiddlewarePanicOnNilHandler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when passing nil handler")
		}
	}()

	_ = Or("test-panic", 10, nil)
}

func TestMiddlewarePriorityAndName(t *testing.T) {
	middleware := Or("test-priority", 25,
		func(c *reqresp.Context) error {
			return nil
		},
	)

	// Check that middleware implements the priority interface
	if m, ok := middleware.(interface{ Priority() int }); !ok {
		t.Error("middleware should implement Priority() method")
	} else if priority := m.Priority(); priority != 25 {
		t.Errorf("expected priority 25, got %d", priority)
	}

	// Check that middleware has a name
	if m, ok := middleware.(interface{ Name() string }); !ok {
		t.Error("middleware should implement Name() method")
	} else if name := m.Name(); name != "test-priority" {
		t.Errorf("expected name 'test-priority', got %s", name)
	}
}

func TestMiddlewareDataPassing(t *testing.T) {
	// Test that data can be set by handler
	middleware := Or("test-data", 10,
		func(c *reqresp.Context) error {
			c.Data = map[string]any{"key": "value", "count": 42}
			return nil
		},
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This will be called because handler returns nil (success)
	})

	handler := middleware.Handler(next)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &reqresp.Context{
		Request: req,
	}
	req = req.WithContext(reqresp.SetContext(req.Context(), ctx))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check data directly in context
	if ctx.Data == nil {
		t.Fatal("expected data to be set by handler")
	}

	if val, ok := ctx.Data["key"]; !ok || val != "value" {
		t.Errorf("expected key='value', got %v", val)
	}

	if val, ok := ctx.Data["count"]; !ok || val != 42 {
		t.Errorf("expected count=42, got %v", val)
	}
}

func TestMiddlewareErrorHandling(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	// Test OR middleware with multiple errors
	middleware := Or("test-errors", 10,
		func(c *reqresp.Context) error {
			return err1
		},
		func(c *reqresp.Context) error {
			return err2
		},
	)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware.Handler(next)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &reqresp.Context{
		Request: req,
	}
	req = req.WithContext(reqresp.SetContext(req.Context(), ctx))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Error("expected next NOT to be called when all handlers fail")
	}

	if ctx.Err == nil {
		t.Fatal("expected error in context")
	}

	// Should have the last error
	if ctx.Err != err2 {
		t.Errorf("expected last error %v, got %v", err2, ctx.Err)
	}
}

func TestMiddlewareEmptyHandlers(t *testing.T) {
	// Test both Or and And with empty handlers
	orMiddleware := Or("empty-or", 10)
	andMiddleware := And("empty-and", 20)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	// Test OR middleware with no handlers
	handler := orMiddleware.Handler(next)
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &reqresp.Context{
		Request: req,
	}
	req = req.WithContext(reqresp.SetContext(req.Context(), ctx))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !nextCalled {
		t.Error("expected next to be called for OR middleware with no handlers")
	}

	// Test AND middleware with no handlers
	nextCalled = false
	handler = andMiddleware.Handler(next)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !nextCalled {
		t.Error("expected next to be called for AND middleware with no handlers")
	}
}

func TestMiddlewareIntegration(t *testing.T) {
	// Test that middleware can be used in a chain
	var calls []int

	orMiddleware := Or("or-test", 20,
		func(c *reqresp.Context) error {
			calls = append(calls, 1)
			return errors.New("handler 1 error")
		},
		func(c *reqresp.Context) error {
			calls = append(calls, 2)
			return nil // This should succeed
		},
	)

	andMiddleware := And("and-test", 10,
		func(c *reqresp.Context) error {
			calls = append(calls, 3)
			return nil // This succeeds
		},
		func(c *reqresp.Context) error {
			calls = append(calls, 4)
			return nil // This also succeeds - all succeed, should call next
		},
	)

	// Create a chain with both middlewares
	chain := Middlewares{andMiddleware, orMiddleware}
	chain.Sort() // Sort by priority (10 then 20)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := chain.Handler(next)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &reqresp.Context{
		Request: req,
	}
	req = req.WithContext(reqresp.SetContext(req.Context(), ctx))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// AND middleware executes first (priority 10), both handlers succeed and calls next
	// OR middleware (priority 20) is the next handler, executes its first handler which fails,
	// then executes its second handler which succeeds
	expectedCalls := []int{3, 4, 1, 2}
	if len(calls) != len(expectedCalls) {
		t.Errorf("expected %d calls, got %d: %v", len(expectedCalls), len(calls), calls)
	} else {
		for i, expected := range expectedCalls {
			if calls[i] != expected {
				t.Errorf("call %d: expected %d, got %d", i, expected, calls[i])
			}
		}
	}

	if !nextCalled {
		t.Error("expected next to be called when all handlers succeed")
	}
}

func TestContextHandlerHandler(t *testing.T) {
	tests := []struct {
		name        string
		handler     ContextHandler
		expectError bool
		expectNext  bool
	}{
		{
			name: "handler returns nil error",
			handler: func(c *reqresp.Context) error {
				c.Data = map[string]any{"test": "success"}
				return nil
			},
			expectError: false,
			expectNext:  true,
		},
		{
			name: "handler returns error",
			handler: func(c *reqresp.Context) error {
				return errors.New("handler error")
			},
			expectError: true,
			expectNext:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a ContextHandler
			ch := tt.handler

			// Test that it implements Middleware interface
			var m Middleware = ch
			_ = m // Use the variable to avoid unused variable warning

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			})

			// Get the handler from the middleware
			handler := ch.Handler(next)

			req := httptest.NewRequest("GET", "/test", nil)
			ctx := &reqresp.Context{
				Request: req,
			}
			req = req.WithContext(reqresp.SetContext(req.Context(), ctx))

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			// Check if next was called
			if nextCalled != tt.expectNext {
				t.Errorf("next called = %v, want %v", nextCalled, tt.expectNext)
			}

			// Check if error was set
			if tt.expectError && ctx.Err == nil {
				t.Error("expected error in context, got nil")
			} else if !tt.expectError && ctx.Err != nil {
				t.Errorf("unexpected error in context: %v", ctx.Err)
			}

			// For the success case, check if data was set
			if !tt.expectError && tt.name == "handler returns nil error" {
				if ctx.Data == nil {
					t.Error("expected data to be set by handler")
				} else if val, ok := ctx.Data["test"]; !ok || val != "success" {
					t.Errorf("expected test='success', got %v", val)
				}
			}
		})
	}
}

func TestContextHandlerHandlerMissingContext(t *testing.T) {
	// Test ContextHandler.Handler when context is missing
	ch := ContextHandler(func(c *reqresp.Context) error {
		return nil
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called when context is missing")
	})

	handler := ch.Handler(next)

	req := httptest.NewRequest("GET", "/test", nil)
	// Don't add context to request

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	expectedBody := "missing reqresp.Context"
	if body := rec.Body.String(); body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}

func TestContextHandlerAsMiddleware(t *testing.T) {
	// Test that ContextHandler can be used as a Middleware in a chain
	var calls []string

	// Create ContextHandlers
	ch1 := ContextHandler(func(c *reqresp.Context) error {
		calls = append(calls, "ch1")
		return nil
	})

	ch2 := ContextHandler(func(c *reqresp.Context) error {
		calls = append(calls, "ch2")
		return errors.New("ch2 error")
	})

	ch3 := ContextHandler(func(c *reqresp.Context) error {
		calls = append(calls, "ch3")
		return nil
	})

	// Create middlewares from ContextHandlers
	m1 := ch1
	m2 := ch2
	m3 := ch3

	// Create a chain of middlewares
	chain := Middlewares{m1, m2, m3}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		calls = append(calls, "next")
	})

	handler := chain.Handler(next)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &reqresp.Context{
		Request: req,
	}
	req = req.WithContext(reqresp.SetContext(req.Context(), ctx))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// ch1 should succeed and call ch2
	// ch2 should fail and stop the chain
	// ch3 should not be called
	// next should not be called
	expectedCalls := []string{"ch1", "ch2"}
	if len(calls) != len(expectedCalls) {
		t.Errorf("expected %d calls, got %d: %v", len(expectedCalls), len(calls), calls)
	} else {
		for i, expected := range expectedCalls {
			if calls[i] != expected {
				t.Errorf("call %d: expected %s, got %s", i, expected, calls[i])
			}
		}
	}

	if nextCalled {
		t.Error("expected next NOT to be called when ch2 fails")
	}

	if ctx.Err == nil {
		t.Error("expected error in context from ch2")
	} else if ctx.Err.Error() != "ch2 error" {
		t.Errorf("expected error 'ch2 error', got %v", ctx.Err)
	}
}

func TestContextHandlerWithOrMiddleware(t *testing.T) {
	// Test that ContextHandler works correctly with Or middleware
	var calls []int

	ch1 := ContextHandler(func(c *reqresp.Context) error {
		calls = append(calls, 1)
		return errors.New("ch1 error")
	})

	ch2 := ContextHandler(func(c *reqresp.Context) error {
		calls = append(calls, 2)
		return nil // Success
	})

	// Create Or middleware using ContextHandlers
	orMiddleware := Or("test-or", 10, ch1, ch2)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		calls = append(calls, 3)
	})

	handler := orMiddleware.Handler(next)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &reqresp.Context{
		Request: req,
	}
	req = req.WithContext(reqresp.SetContext(req.Context(), ctx))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// ch1 should be called and fail
	// ch2 should be called and succeed
	// next should be called
	expectedCalls := []int{1, 2, 3}
	if len(calls) != len(expectedCalls) {
		t.Errorf("expected %d calls, got %d: %v", len(expectedCalls), len(calls), calls)
	} else {
		for i, expected := range expectedCalls {
			if calls[i] != expected {
				t.Errorf("call %d: expected %d, got %d", i, expected, calls[i])
			}
		}
	}

	if !nextCalled {
		t.Error("expected next to be called when ch2 succeeds")
	}

	if ctx.Err != nil {
		t.Errorf("unexpected error in context: %v", ctx.Err)
	}
}

func TestContextHandlerWithAndMiddleware(t *testing.T) {
	// Test that ContextHandler works correctly with And middleware
	var calls []int

	ch1 := ContextHandler(func(c *reqresp.Context) error {
		calls = append(calls, 1)
		return nil // Success
	})

	ch2 := ContextHandler(func(c *reqresp.Context) error {
		calls = append(calls, 2)
		return errors.New("ch2 error") // Fail
	})

	// Create And middleware using ContextHandlers
	andMiddleware := And("test-and", 10, ch1, ch2)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		calls = append(calls, 3)
	})

	handler := andMiddleware.Handler(next)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &reqresp.Context{
		Request: req,
	}
	req = req.WithContext(reqresp.SetContext(req.Context(), ctx))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// ch1 should be called and succeed
	// ch2 should be called and fail
	// next should NOT be called
	expectedCalls := []int{1, 2}
	if len(calls) != len(expectedCalls) {
		t.Errorf("expected %d calls, got %d: %v", len(expectedCalls), len(calls), calls)
	} else {
		for i, expected := range expectedCalls {
			if calls[i] != expected {
				t.Errorf("call %d: expected %d, got %d", i, expected, calls[i])
			}
		}
	}

	if nextCalled {
		t.Error("expected next NOT to be called when ch2 fails")
	}

	if ctx.Err == nil {
		t.Error("expected error in context from ch2")
	} else if ctx.Err.Error() != "ch2 error" {
		t.Errorf("expected error 'ch2 error', got %v", ctx.Err)
	}
}
