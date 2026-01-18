// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package core

import (
	"encoding/json"
	"fmt"
)

// ToResource converts any struct or map to a Resource via JSON marshaling.
// This is useful for converting provider-specific types to the generic Resource type.
//
// Note: JSON marshaling converts all numbers to float64 and may lose type information.
// Use this function when you need a generic map representation of structured data.
func ToResource(item any) (Resource, error) {
	if item == nil {
		return nil, fmt.Errorf("cannot convert nil to Resource")
	}

	data, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}

	var result Resource
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return result, nil
}

// ToResourceWithFields converts any struct or map to a Resource and adds/overrides
// fields from the provided map. This is useful when you need to add computed fields
// or metadata to a resource.
func ToResourceWithFields(item any, fields map[string]any) (Resource, error) {
	result, err := ToResource(item)
	if err != nil {
		return nil, err
	}

	for k, v := range fields {
		result[k] = v
	}

	return result, nil
}

// MustToResource is like ToResource but panics on error.
// Use this only when you are certain the input is valid and marshalable.
func MustToResource(item any) Resource {
	result, err := ToResource(item)
	if err != nil {
		panic(fmt.Sprintf("MustToResource: %v", err))
	}
	return result
}
