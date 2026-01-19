// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterByProvider(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "aws-policy"},
			Require:  []Requirement{{Provider: "aws"}},
		},
		{
			Metadata: Metadata{Name: "azure-policy"},
			Require:  []Requirement{{Provider: "azure"}},
		},
		{
			Metadata: Metadata{Name: "multi-policy"},
			Require:  []Requirement{{Provider: "aws"}, {Provider: "azure"}},
		},
		{
			Metadata: Metadata{Name: "no-require-policy"},
			Require:  nil,
		},
	}

	t.Run("filter for aws", func(t *testing.T) {
		filtered := FilterByProvider(policies, "aws", "")
		assert.Len(t, filtered, 3) // aws-policy, multi-policy, no-require-policy

		names := make(map[string]bool)
		for _, p := range filtered {
			names[p.Metadata.Name] = true
		}
		assert.True(t, names["aws-policy"])
		assert.True(t, names["multi-policy"])
		assert.True(t, names["no-require-policy"])
		assert.False(t, names["azure-policy"])
	})

	t.Run("filter for azure", func(t *testing.T) {
		filtered := FilterByProvider(policies, "azure", "")
		assert.Len(t, filtered, 3) // azure-policy, multi-policy, no-require-policy

		names := make(map[string]bool)
		for _, p := range filtered {
			names[p.Metadata.Name] = true
		}
		assert.True(t, names["azure-policy"])
		assert.True(t, names["multi-policy"])
		assert.True(t, names["no-require-policy"])
		assert.False(t, names["aws-policy"])
	})

	t.Run("filter with alias", func(t *testing.T) {
		filtered := FilterByProvider(policies, "network", "aws")
		assert.Len(t, filtered, 3) // aws-policy matches alias, multi-policy, no-require-policy
	})

	t.Run("filter for unknown provider", func(t *testing.T) {
		filtered := FilterByProvider(policies, "unknown", "")
		assert.Len(t, filtered, 1) // Only no-require-policy
		assert.Equal(t, "no-require-policy", filtered[0].Metadata.Name)
	})

	t.Run("empty policies", func(t *testing.T) {
		filtered := FilterByProvider([]Policy{}, "aws", "")
		assert.Len(t, filtered, 0)
	})
}
