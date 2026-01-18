// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package core

import (
	"errors"
	"fmt"
)

// ProviderError represents an error that occurred during provider operations.
// It includes context about which provider, resource, and operation failed.
type ProviderError struct {
	Provider  string // Provider name (e.g., "aws", "azure")
	Resource  string // Resource type (e.g., "s3_bucket", "vm")
	Operation string // Operation that failed (e.g., "list", "fetch", "describe")
	Err       error  // Underlying error
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s: %s %s failed: %s", e.Provider, e.Resource, e.Operation, e.Err.Error())
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new ProviderError with the given context.
func NewProviderError(provider, resource, operation string, err error) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Resource:  resource,
		Operation: operation,
		Err:       err,
	}
}

// WrapError is a convenience function that creates a ProviderError.
// It's useful for inline error wrapping in provider code.
func WrapError(provider, resource, operation string, err error) error {
	return NewProviderError(provider, resource, operation, err)
}

// ResourceError represents an error related to a specific resource instance.
// Use this when you have a specific resource ID.
type ResourceError struct {
	Resource   string // Resource type (e.g., "aws_s3_bucket")
	ResourceID string // Specific resource identifier
	Operation  string // Operation that failed
	Err        error  // Underlying error
}

// Error implements the error interface.
func (e *ResourceError) Error() string {
	if e.ResourceID != "" {
		return fmt.Sprintf("%s[%s] %s: %s", e.Resource, e.ResourceID, e.Operation, e.Err.Error())
	}
	return fmt.Sprintf("%s %s: %s", e.Resource, e.Operation, e.Err.Error())
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *ResourceError) Unwrap() error {
	return e.Err
}

// NewResourceError creates a new ResourceError with the given context.
func NewResourceError(resource, resourceID, operation string, err error) *ResourceError {
	return &ResourceError{
		Resource:   resource,
		ResourceID: resourceID,
		Operation:  operation,
		Err:        err,
	}
}

// IsProviderError checks if the error (or any wrapped error) is a ProviderError.
func IsProviderError(err error) bool {
	var pe *ProviderError
	return errors.As(err, &pe)
}

// GetProviderFromError extracts the provider name from a ProviderError.
// Returns empty string and false if the error is not a ProviderError.
func GetProviderFromError(err error) (string, bool) {
	var pe *ProviderError
	if errors.As(err, &pe) {
		return pe.Provider, true
	}
	return "", false
}
