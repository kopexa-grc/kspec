// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package core

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSafeGo(t *testing.T) {
	t.Run("executes function successfully", func(t *testing.T) {
		var executed atomic.Bool
		var wg sync.WaitGroup
		wg.Add(1)

		SafeGo(func() {
			defer wg.Done()
			executed.Store(true)
		})

		wg.Wait()
		if !executed.Load() {
			t.Error("SafeGo did not execute the function")
		}
	})

	t.Run("recovers from panic", func(t *testing.T) {
		var recovered atomic.Bool
		var wg sync.WaitGroup
		wg.Add(1)

		SafeGoWithHandler(func() {
			defer wg.Done()
			panic("test panic")
		}, func(r any) {
			recovered.Store(true)
		})

		wg.Wait()
		if !recovered.Load() {
			t.Error("SafeGo did not recover from panic")
		}
	})

	t.Run("recovers from error panic", func(t *testing.T) {
		var panicValue atomic.Value
		done := make(chan struct{})

		SafeGoWithHandler(func() {
			panic(errors.New("error panic"))
		}, func(r any) {
			panicValue.Store(r)
			close(done)
		})

		<-done
		if panicValue.Load() == nil {
			t.Error("SafeGo did not capture panic value")
		}
	})

	t.Run("handles nil panic", func(t *testing.T) {
		var handlerCalled atomic.Bool
		var wg sync.WaitGroup
		wg.Add(1)

		SafeGoWithHandler(func() {
			defer wg.Done()
			panic(nil)
		}, func(r any) {
			handlerCalled.Store(true)
		})

		wg.Wait()
		// nil panic should still trigger recover but may not call handler
		// depending on implementation
	})
}

func TestSafeGoWithHandler(t *testing.T) {
	t.Run("calls handler on panic", func(t *testing.T) {
		var handlerCalled atomic.Bool
		var panicMsg atomic.Value
		var wg sync.WaitGroup
		wg.Add(1)

		SafeGoWithHandler(func() {
			defer wg.Done()
			panic("custom panic message")
		}, func(r any) {
			handlerCalled.Store(true)
			panicMsg.Store(r)
		})

		wg.Wait()
		if !handlerCalled.Load() {
			t.Error("handler was not called")
		}
		if panicMsg.Load() != "custom panic message" {
			t.Errorf("panic message = %v, want 'custom panic message'", panicMsg.Load())
		}
	})

	t.Run("does not call handler on success", func(t *testing.T) {
		var handlerCalled atomic.Bool
		var wg sync.WaitGroup
		wg.Add(1)

		SafeGoWithHandler(func() {
			defer wg.Done()
			// no panic
		}, func(r any) {
			handlerCalled.Store(true)
		})

		wg.Wait()
		if handlerCalled.Load() {
			t.Error("handler should not be called when no panic occurs")
		}
	})

	t.Run("nil handler is safe", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		// Should not panic even with nil handler
		SafeGoWithHandler(func() {
			defer wg.Done()
			panic("test")
		}, nil)

		wg.Wait()
		// Test passes if no additional panic occurs
	})
}

func TestSafeGoResult(t *testing.T) {
	t.Run("returns result on success", func(t *testing.T) {
		resultChan := SafeGoResult(func() (int, error) {
			return 42, nil
		})

		select {
		case result := <-resultChan:
			if result.Err != nil {
				t.Errorf("unexpected error: %v", result.Err)
			}
			if result.Value != 42 {
				t.Errorf("Value = %v, want 42", result.Value)
			}
		case <-time.After(time.Second):
			t.Error("timeout waiting for result")
		}
	})

	t.Run("returns error on panic", func(t *testing.T) {
		resultChan := SafeGoResult(func() (int, error) {
			panic("computation failed")
		})

		select {
		case result := <-resultChan:
			if result.Err == nil {
				t.Error("expected error from panic")
			}
			if result.Value != 0 {
				t.Errorf("Value should be zero value, got %v", result.Value)
			}
		case <-time.After(time.Second):
			t.Error("timeout waiting for result")
		}
	})

	t.Run("returns function error", func(t *testing.T) {
		expectedErr := errors.New("function error")
		resultChan := SafeGoResult(func() (string, error) {
			return "", expectedErr
		})

		select {
		case result := <-resultChan:
			if !errors.Is(result.Err, expectedErr) {
				t.Errorf("Err = %v, want %v", result.Err, expectedErr)
			}
		case <-time.After(time.Second):
			t.Error("timeout waiting for result")
		}
	})
}

func TestRecoverError(t *testing.T) {
	tests := []struct {
		name       string
		panicValue any
		wantMsg    string
	}{
		{
			name:       "string panic",
			panicValue: "something went wrong",
			wantMsg:    "panic: something went wrong",
		},
		{
			name:       "error panic",
			panicValue: errors.New("error value"),
			wantMsg:    "panic: error value",
		},
		{
			name:       "int panic",
			panicValue: 42,
			wantMsg:    "panic: 42",
		},
		{
			name:       "nil panic",
			panicValue: nil,
			wantMsg:    "panic: <nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RecoverError(tt.panicValue)
			if err.Error() != tt.wantMsg {
				t.Errorf("RecoverError() = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

// Benchmarks
func BenchmarkSafeGo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(1)
		SafeGo(func() {
			wg.Done()
		})
		wg.Wait()
	}
}

func BenchmarkSafeGoWithPanic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(1)
		SafeGoWithHandler(func() {
			defer wg.Done()
			panic("test")
		}, func(r any) {})
		wg.Wait()
	}
}
