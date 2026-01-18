// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package azure

import (
	"context"
	"errors"
	"testing"
)

func TestToResourceMap(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
		check   func(map[string]interface{}) bool
	}{
		{
			name: "simple struct",
			input: &struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				ID:   "test-id",
				Name: "test-name",
			},
			wantErr: false,
			check: func(m map[string]interface{}) bool {
				return m["id"] == "test-id" && m["name"] == "test-name"
			},
		},
		{
			name: "nested struct",
			input: &struct {
				ID         string `json:"id"`
				Properties struct {
					Enabled bool `json:"enabled"`
				} `json:"properties"`
			}{
				ID: "nested-id",
				Properties: struct {
					Enabled bool `json:"enabled"`
				}{
					Enabled: true,
				},
			},
			wantErr: false,
			check: func(m map[string]interface{}) bool {
				props, ok := m["properties"].(map[string]interface{})
				return ok && props["enabled"] == true
			},
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true, // nil inputs should return error
			check: func(m map[string]interface{}) bool {
				return m == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toResourceMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("toResourceMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(got) {
				t.Errorf("toResourceMap() check failed, got %v", got)
			}
		})
	}
}

// Mock pager for testing
type mockPager[T any] struct {
	pages   []T
	current int
	err     error
}

func (m *mockPager[T]) More() bool {
	return m.current < len(m.pages)
}

func (m *mockPager[T]) NextPage(ctx context.Context) (T, error) {
	if m.err != nil {
		var zero T
		return zero, m.err
	}
	page := m.pages[m.current]
	m.current++
	return page, nil
}

type testPage struct {
	Items []*testItem
}

type testItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestFetchWithPager(t *testing.T) {
	t.Run("fetches all pages", func(t *testing.T) {
		pager := &mockPager[testPage]{
			pages: []testPage{
				{Items: []*testItem{{ID: "1", Name: "first"}}},
				{Items: []*testItem{{ID: "2", Name: "second"}, {ID: "3", Name: "third"}}},
			},
		}

		extractor := func(page testPage) []*testItem {
			return page.Items
		}

		resources, err := fetchWithPager(context.Background(), pager, extractor, "test_resource")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resources) != 3 {
			t.Errorf("expected 3 resources, got %d", len(resources))
		}
	})

	t.Run("handles empty pages", func(t *testing.T) {
		pager := &mockPager[testPage]{
			pages: []testPage{},
		}

		extractor := func(page testPage) []*testItem {
			return page.Items
		}

		resources, err := fetchWithPager(context.Background(), pager, extractor, "test_resource")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resources) != 0 {
			t.Errorf("expected 0 resources, got %d", len(resources))
		}
	})

	t.Run("returns error on pager failure", func(t *testing.T) {
		expectedErr := errors.New("pager error")
		pager := &mockPager[testPage]{
			pages: []testPage{{}},
			err:   expectedErr,
		}

		extractor := func(page testPage) []*testItem {
			return page.Items
		}

		_, err := fetchWithPager(context.Background(), pager, extractor, "test_resource")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func BenchmarkToResourceMap(b *testing.B) {
	item := &struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Location   string `json:"location"`
		Properties struct {
			Enabled     bool   `json:"enabled"`
			Description string `json:"description"`
		} `json:"properties"`
	}{
		ID:       "bench-id",
		Name:     "bench-name",
		Location: "eastus",
		Properties: struct {
			Enabled     bool   `json:"enabled"`
			Description string `json:"description"`
		}{
			Enabled:     true,
			Description: "benchmark test",
		},
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = toResourceMap(item)
	}
}
