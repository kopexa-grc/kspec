// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package core

import (
	"errors"
	"fmt"
	"testing"
)

func TestProviderError(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		resource  string
		operation string
		err       error
		wantMsg   string
	}{
		{
			name:      "basic error",
			provider:  "aws",
			resource:  "s3_bucket",
			operation: "list",
			err:       errors.New("access denied"),
			wantMsg:   "aws: s3_bucket list failed: access denied",
		},
		{
			name:      "nested error",
			provider:  "azure",
			resource:  "storage_account",
			operation: "fetch",
			err:       fmt.Errorf("API error: %w", errors.New("timeout")),
			wantMsg:   "azure: storage_account fetch failed: API error: timeout",
		},
		{
			name:      "empty operation",
			provider:  "github",
			resource:  "repository",
			operation: "",
			err:       errors.New("not found"),
			wantMsg:   "github: repository  failed: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewProviderError(tt.provider, tt.resource, tt.operation, tt.err)

			// Test Error() message
			if got := err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}

			// Test Unwrap()
			if !errors.Is(err, tt.err) {
				t.Errorf("Unwrap() should return original error")
			}

			// Test fields
			var pe *ProviderError
			if !errors.As(err, &pe) {
				t.Fatal("errors.As() should work with *ProviderError")
			}
			if pe.Provider != tt.provider {
				t.Errorf("Provider = %q, want %q", pe.Provider, tt.provider)
			}
			if pe.Resource != tt.resource {
				t.Errorf("Resource = %q, want %q", pe.Resource, tt.resource)
			}
			if pe.Operation != tt.operation {
				t.Errorf("Operation = %q, want %q", pe.Operation, tt.operation)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("connection refused")

	wrapped := WrapError("aws", "ec2_instance", "describe", originalErr)

	// Check message format
	expected := "aws: ec2_instance describe failed: connection refused"
	if wrapped.Error() != expected {
		t.Errorf("WrapError() = %q, want %q", wrapped.Error(), expected)
	}

	// Check unwrapping
	if !errors.Is(wrapped, originalErr) {
		t.Error("WrapError() should preserve original error for errors.Is()")
	}
}

func TestResourceError(t *testing.T) {
	tests := []struct {
		name       string
		resource   string
		resourceID string
		operation  string
		err        error
		wantMsg    string
	}{
		{
			name:       "with resource ID",
			resource:   "aws_s3_bucket",
			resourceID: "my-bucket-123",
			operation:  "get_encryption",
			err:        errors.New("permission denied"),
			wantMsg:    "aws_s3_bucket[my-bucket-123] get_encryption: permission denied",
		},
		{
			name:       "without resource ID",
			resource:   "aws_iam_user",
			resourceID: "",
			operation:  "list",
			err:        errors.New("rate limited"),
			wantMsg:    "aws_iam_user list: rate limited",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewResourceError(tt.resource, tt.resourceID, tt.operation, tt.err)

			if got := err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}

			if !errors.Is(err, tt.err) {
				t.Error("Unwrap() should return original error")
			}
		})
	}
}

func TestIsProviderError(t *testing.T) {
	providerErr := NewProviderError("aws", "s3", "list", errors.New("test"))
	wrappedProviderErr := fmt.Errorf("outer: %w", providerErr)
	regularErr := errors.New("regular error")

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"provider error", providerErr, true},
		{"wrapped provider error", wrappedProviderErr, true},
		{"regular error", regularErr, false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsProviderError(tt.err); got != tt.want {
				t.Errorf("IsProviderError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProviderFromError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantProvider string
		wantOK       bool
	}{
		{
			name:         "provider error",
			err:          NewProviderError("aws", "s3", "list", errors.New("test")),
			wantProvider: "aws",
			wantOK:       true,
		},
		{
			name:         "wrapped provider error",
			err:          fmt.Errorf("outer: %w", NewProviderError("azure", "vm", "get", errors.New("test"))),
			wantProvider: "azure",
			wantOK:       true,
		},
		{
			name:         "regular error",
			err:          errors.New("regular"),
			wantProvider: "",
			wantOK:       false,
		},
		{
			name:         "nil error",
			err:          nil,
			wantProvider: "",
			wantOK:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProvider, gotOK := GetProviderFromError(tt.err)
			if gotProvider != tt.wantProvider || gotOK != tt.wantOK {
				t.Errorf("GetProviderFromError() = (%q, %v), want (%q, %v)",
					gotProvider, gotOK, tt.wantProvider, tt.wantOK)
			}
		})
	}
}

// Benchmark error creation
func BenchmarkNewProviderError(b *testing.B) {
	originalErr := errors.New("test error")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewProviderError("aws", "s3_bucket", "list", originalErr)
	}
}

func BenchmarkProviderErrorMessage(b *testing.B) {
	err := NewProviderError("aws", "s3_bucket", "list", errors.New("test"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
