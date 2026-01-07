package core

import (
	"encoding/json"
	"testing"
)

func TestToResource(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    Resource
		wantErr bool
	}{
		{
			name: "simple struct",
			input: struct {
				Name  string `json:"name"`
				Count int    `json:"count"`
			}{
				Name:  "test",
				Count: 42,
			},
			want: Resource{
				"name":  "test",
				"count": float64(42), // JSON numbers become float64
			},
			wantErr: false,
		},
		{
			name: "nested struct",
			input: struct {
				ID     string `json:"id"`
				Config struct {
					Enabled bool `json:"enabled"`
				} `json:"config"`
			}{
				ID: "abc123",
				Config: struct {
					Enabled bool `json:"enabled"`
				}{
					Enabled: true,
				},
			},
			want: Resource{
				"id": "abc123",
				"config": map[string]any{
					"enabled": true,
				},
			},
			wantErr: false,
		},
		{
			name: "map input",
			input: map[string]any{
				"key": "value",
				"num": 123,
			},
			want: Resource{
				"key": "value",
				"num": float64(123),
			},
			wantErr: false,
		},
		{
			name: "slice of strings",
			input: struct {
				Tags []string `json:"tags"`
			}{
				Tags: []string{"a", "b", "c"},
			},
			want: Resource{
				"tags": []any{"a", "b", "c"},
			},
			wantErr: false,
		},
		{
			name:  "nil input",
			input: nil,
			want:  Resource{
				// nil marshals to "null" which unmarshals to nil map
			},
			wantErr: true, // nil cannot be unmarshaled to map
		},
		{
			name:    "channel - unmarshalable",
			input:   make(chan int),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "function - unmarshalable",
			input:   func() {},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty struct",
			input:   struct{}{},
			want:    Resource{},
			wantErr: false,
		},
		{
			name: "pointer to struct",
			input: &struct {
				Value string `json:"value"`
			}{
				Value: "ptr",
			},
			want: Resource{
				"value": "ptr",
			},
			wantErr: false,
		},
		{
			name: "struct with omitempty",
			input: struct {
				Name  string `json:"name"`
				Empty string `json:"empty,omitempty"`
			}{
				Name:  "test",
				Empty: "",
			},
			want: Resource{
				"name": "test",
				// "empty" should be omitted
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToResource(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !resourceEqual(got, tt.want) {
				t.Errorf("ToResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToResourceWithFields(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		fields  map[string]any
		want    Resource
		wantErr bool
	}{
		{
			name: "add extra fields",
			input: struct {
				Name string `json:"name"`
			}{
				Name: "test",
			},
			fields: map[string]any{
				"id":     "123",
				"active": true,
			},
			want: Resource{
				"name":   "test",
				"id":     "123",
				"active": true,
			},
			wantErr: false,
		},
		{
			name: "override existing field",
			input: struct {
				Name string `json:"name"`
			}{
				Name: "original",
			},
			fields: map[string]any{
				"name": "overridden",
			},
			want: Resource{
				"name": "overridden",
			},
			wantErr: false,
		},
		{
			name: "nil fields",
			input: struct {
				Name string `json:"name"`
			}{
				Name: "test",
			},
			fields: nil,
			want: Resource{
				"name": "test",
			},
			wantErr: false,
		},
		{
			name: "empty fields",
			input: struct {
				Name string `json:"name"`
			}{
				Name: "test",
			},
			fields: map[string]any{},
			want: Resource{
				"name": "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToResourceWithFields(tt.input, tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToResourceWithFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !resourceEqual(got, tt.want) {
				t.Errorf("ToResourceWithFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustToResource(t *testing.T) {
	// Test successful case
	t.Run("success", func(t *testing.T) {
		input := struct {
			Name string `json:"name"`
		}{Name: "test"}

		got := MustToResource(input)
		if got["name"] != "test" {
			t.Errorf("MustToResource() = %v, want name=test", got)
		}
	})

	// Test panic case
	t.Run("panics on error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustToResource() did not panic on unmarshalable input")
			}
		}()

		MustToResource(make(chan int))
	})
}

// resourceEqual compares two Resources for equality
func resourceEqual(a, b Resource) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if !valueEqual(v, bv) {
			return false
		}
	}
	return true
}

// valueEqual compares two values for equality (handles nested maps and slices)
func valueEqual(a, b any) bool {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok {
			return false
		}
		if len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !valueEqual(v, bv[k]) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok {
			return false
		}
		if len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !valueEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// Fuzz test for ToResource
func FuzzToResource(f *testing.F) {
	// Add seed corpus
	f.Add(`{"name":"test","count":42}`)
	f.Add(`{"nested":{"key":"value"}}`)
	f.Add(`{"array":[1,2,3]}`)
	f.Add(`{"empty":{}}`)
	f.Add(`{"mixed":{"str":"hello","num":123,"bool":true,"null":null}}`)
	f.Add(`{}`)

	f.Fuzz(func(t *testing.T, jsonStr string) {
		// Parse JSON string to map first
		var input map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
			// Invalid JSON, skip
			return
		}

		// Convert to Resource
		result, err := ToResource(input)
		if err != nil {
			// Some inputs may fail, that's ok
			return
		}

		// Verify result is valid
		if result == nil && len(input) > 0 {
			t.Error("ToResource returned nil for non-empty input")
		}

		// Verify roundtrip: input should equal result
		if len(input) != len(result) {
			t.Errorf("Length mismatch: input=%d, result=%d", len(input), len(result))
		}
	})
}
