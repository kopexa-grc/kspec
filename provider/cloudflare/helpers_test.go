package cloudflare

import (
	"testing"

	"github.com/kopexa-grc/kspec/core"
)

func TestResolveAccountID(t *testing.T) {
	tests := []struct {
		name              string
		resourceAccountID string
		asset             core.Asset
		resourceName      string
		want              string
		wantErr           bool
	}{
		{
			name:              "from resource config",
			resourceAccountID: "acc-123",
			asset:             core.Asset{},
			resourceName:      "dns_record",
			want:              "acc-123",
			wantErr:           false,
		},
		{
			name:              "from asset config",
			resourceAccountID: "",
			asset: core.Asset{
				Config: map[string]string{"account_id": "acc-456"},
			},
			resourceName: "dns_record",
			want:         "acc-456",
			wantErr:      false,
		},
		{
			name:              "resource takes precedence",
			resourceAccountID: "acc-123",
			asset: core.Asset{
				Config: map[string]string{"account_id": "acc-456"},
			},
			resourceName: "dns_record",
			want:         "acc-123",
			wantErr:      false,
		},
		{
			name:              "missing account_id",
			resourceAccountID: "",
			asset:             core.Asset{},
			resourceName:      "dns_record",
			want:              "",
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveAccountID(tt.resourceAccountID, tt.asset, tt.resourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveAccountID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveAccountID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItemsToResources(t *testing.T) {
	type testItem struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	t.Run("converts items to resources", func(t *testing.T) {
		items := []testItem{
			{ID: "1", Name: "first"},
			{ID: "2", Name: "second"},
		}

		resources := itemsToResources(items, "acc-123", nil)

		if len(resources) != 2 {
			t.Fatalf("expected 2 resources, got %d", len(resources))
		}

		// Check first resource
		if resources[0]["id"] != "1" {
			t.Errorf("expected id=1, got %v", resources[0]["id"])
		}
		if resources[0]["name"] != "first" {
			t.Errorf("expected name=first, got %v", resources[0]["name"])
		}
		if resources[0]["account_id"] != "acc-123" {
			t.Errorf("expected account_id=acc-123, got %v", resources[0]["account_id"])
		}

		// Check second resource
		if resources[1]["id"] != "2" {
			t.Errorf("expected id=2, got %v", resources[1]["id"])
		}
	})

	t.Run("applies addFields callback", func(t *testing.T) {
		items := []testItem{{ID: "1", Name: "test"}}

		resources := itemsToResources(items, "acc-123", func(m map[string]interface{}) {
			m["extra_field"] = "extra_value"
			m["computed"] = true
		})

		if len(resources) != 1 {
			t.Fatalf("expected 1 resource, got %d", len(resources))
		}

		if resources[0]["extra_field"] != "extra_value" {
			t.Errorf("expected extra_field=extra_value, got %v", resources[0]["extra_field"])
		}
		if resources[0]["computed"] != true {
			t.Errorf("expected computed=true, got %v", resources[0]["computed"])
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		items := []testItem{}

		resources := itemsToResources(items, "acc-123", nil)

		if len(resources) != 0 {
			t.Errorf("expected 0 resources, got %d", len(resources))
		}
	})

	t.Run("handles nested structs", func(t *testing.T) {
		type nestedItem struct {
			ID     string `json:"id"`
			Config struct {
				Enabled bool   `json:"enabled"`
				Mode    string `json:"mode"`
			} `json:"config"`
		}

		items := []nestedItem{
			{
				ID: "1",
				Config: struct {
					Enabled bool   `json:"enabled"`
					Mode    string `json:"mode"`
				}{
					Enabled: true,
					Mode:    "strict",
				},
			},
		}

		resources := itemsToResources(items, "acc-123", nil)

		if len(resources) != 1 {
			t.Fatalf("expected 1 resource, got %d", len(resources))
		}

		config, ok := resources[0]["config"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected config to be map, got %T", resources[0]["config"])
		}
		if config["enabled"] != true {
			t.Errorf("expected config.enabled=true, got %v", config["enabled"])
		}
		if config["mode"] != "strict" {
			t.Errorf("expected config.mode=strict, got %v", config["mode"])
		}
	})
}

// Benchmark to ensure refactoring doesn't regress performance
func BenchmarkItemsToResources(b *testing.B) {
	type benchItem struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
		Count       int    `json:"count"`
	}

	items := make([]benchItem, 100)
	for i := range items {
		items[i] = benchItem{
			ID:          "id-" + string(rune(i)),
			Name:        "name-" + string(rune(i)),
			Description: "description text here",
			Enabled:     true,
			Count:       i,
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_ = itemsToResources(items, "acc-123", nil)
	}
}
