// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

// Package ptr provides generic pointer utility functions for safe pointer operations.
package ptr

import (
	"fmt"
	"reflect"
)

// To returns a pointer to the given value.
func To[T any](v T) *T {
	return &v
}

// Deref safely dereferences a pointer, returning the default value if nil.
func Deref[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}

// Equal compares two pointers for equality.
func Equal[T comparable](a, b *T) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}
	return *a == *b
}

// AllPtrFieldsNil returns true if all pointer fields in a struct are nil.
func AllPtrFieldsNil(obj interface{}) bool {
	v := reflect.ValueOf(obj)
	if !v.IsValid() {
		panic(fmt.Sprintf("reflect.ValueOf() produced a non-valid Value for %#v", obj))
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.Ptr && !v.Field(i).IsNil() {
			return false
		}
	}
	return true
}
