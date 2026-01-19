// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package provider

import (
	"context"
	"sync"
	"testing"

	"github.com/kopexa-grc/kspec/core"
)

// mockProvider implements core.Provider for testing.
type mockProvider struct{}

func (m *mockProvider) Name() string { return "mock" }
func (m *mockProvider) Connect(_ context.Context, _ map[string]string) (core.Connection, error) {
	return nil, nil
}

func mockFactory() core.Provider {
	return &mockProvider{}
}

// resetRegistry clears the global registry state for testing.
func resetRegistry() {
	mu.Lock()
	defer mu.Unlock()
	registeredProvider = make(map[string]*ProviderDefinition)
	aliases = make(map[string]string)
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name      string
		def       *ProviderDefinition
		wantPanic bool
		panicMsg  string
	}{
		{
			name:      "nil definition panics",
			def:       nil,
			wantPanic: true,
			panicMsg:  "provider definition must not be nil",
		},
		{
			name: "empty name panics",
			def: &ProviderDefinition{
				Name:    "",
				Factory: mockFactory,
			},
			wantPanic: true,
			panicMsg:  "provider definition must have a name",
		},
		{
			name: "nil factory panics",
			def: &ProviderDefinition{
				Name:    "test",
				Factory: nil,
			},
			wantPanic: true,
			panicMsg:  "provider \"test\" must have a non-nil factory",
		},
		{
			name: "valid registration succeeds",
			def: &ProviderDefinition{
				Name:        "valid",
				Description: "Valid provider",
				Factory:     mockFactory,
			},
			wantPanic: false,
		},
		{
			name: "registration with aliases succeeds",
			def: &ProviderDefinition{
				Name:    "with-aliases",
				Aliases: []string{"alias1", "alias2"},
				Factory: mockFactory,
			},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetRegistry()

			if tt.wantPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Error("expected panic, but didn't get one")
					}
					if msg, ok := r.(string); ok && msg != tt.panicMsg {
						t.Errorf("panic message = %q, want %q", msg, tt.panicMsg)
					}
				}()
			}

			Register(tt.def)

			if !tt.wantPanic {
				// Verify registration
				def, ok := Get(tt.def.Name)
				if !ok {
					t.Errorf("provider %q not found after registration", tt.def.Name)
				}
				if def.Name != tt.def.Name {
					t.Errorf("registered provider name = %q, want %q", def.Name, tt.def.Name)
				}
			}
		})
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	resetRegistry()

	def := &ProviderDefinition{
		Name:    "duplicate",
		Factory: mockFactory,
	}

	Register(def)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for duplicate registration")
		}
	}()

	Register(def)
}

func TestRegisterDuplicateAliasPanics(t *testing.T) {
	resetRegistry()

	def1 := &ProviderDefinition{
		Name:    "provider1",
		Aliases: []string{"shared-alias"},
		Factory: mockFactory,
	}
	def2 := &ProviderDefinition{
		Name:    "provider2",
		Aliases: []string{"shared-alias"},
		Factory: mockFactory,
	}

	Register(def1)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for duplicate alias")
		}
	}()

	Register(def2)
}

func TestGet(t *testing.T) {
	resetRegistry()

	def := &ProviderDefinition{
		Name:    "testget",
		Aliases: []string{"tg", "test-get"},
		Factory: mockFactory,
	}
	Register(def)

	tests := []struct {
		name      string
		query     string
		wantFound bool
		wantName  string
	}{
		{
			name:      "get by primary name",
			query:     "testget",
			wantFound: true,
			wantName:  "testget",
		},
		{
			name:      "get by alias",
			query:     "tg",
			wantFound: true,
			wantName:  "testget",
		},
		{
			name:      "get by second alias",
			query:     "test-get",
			wantFound: true,
			wantName:  "testget",
		},
		{
			name:      "get nonexistent returns false",
			query:     "nonexistent",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, ok := Get(tt.query)
			if ok != tt.wantFound {
				t.Errorf("Get(%q) found = %v, want %v", tt.query, ok, tt.wantFound)
			}
			if ok && def.Name != tt.wantName {
				t.Errorf("Get(%q).Name = %q, want %q", tt.query, def.Name, tt.wantName)
			}
		})
	}
}

func TestAll(t *testing.T) {
	resetRegistry()

	// Register providers in non-alphabetical order
	defs := []*ProviderDefinition{
		{Name: "zebra", Factory: mockFactory},
		{Name: "alpha", Factory: mockFactory},
		{Name: "middle", Factory: mockFactory},
	}

	for _, def := range defs {
		Register(def)
	}

	all := All()

	if len(all) != 3 {
		t.Errorf("All() returned %d providers, want 3", len(all))
	}

	// Verify sorted order
	expectedOrder := []string{"alpha", "middle", "zebra"}
	for i, def := range all {
		if def.Name != expectedOrder[i] {
			t.Errorf("All()[%d].Name = %q, want %q", i, def.Name, expectedOrder[i])
		}
	}
}

func TestNames(t *testing.T) {
	resetRegistry()

	defs := []*ProviderDefinition{
		{Name: "provider-c", Aliases: []string{"c"}, Factory: mockFactory},
		{Name: "provider-a", Aliases: []string{"a"}, Factory: mockFactory},
		{Name: "provider-b", Aliases: []string{"b"}, Factory: mockFactory},
	}

	for _, def := range defs {
		Register(def)
	}

	names := Names()

	if len(names) != 3 {
		t.Errorf("Names() returned %d names, want 3", len(names))
	}

	// Verify sorted and no aliases included
	expectedNames := []string{"provider-a", "provider-b", "provider-c"}
	for i, name := range names {
		if name != expectedNames[i] {
			t.Errorf("Names()[%d] = %q, want %q", i, name, expectedNames[i])
		}
	}
}

func TestExists(t *testing.T) {
	resetRegistry()

	Register(&ProviderDefinition{
		Name:    "exists-test",
		Aliases: []string{"et"},
		Factory: mockFactory,
	})

	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{"primary name exists", "exists-test", true},
		{"alias exists", "et", true},
		{"nonexistent", "nope", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Exists(tt.query); got != tt.want {
				t.Errorf("Exists(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestMustGet(t *testing.T) {
	resetRegistry()

	Register(&ProviderDefinition{
		Name:    "must-get-test",
		Factory: mockFactory,
	})

	t.Run("found returns definition", func(t *testing.T) {
		def := MustGet("must-get-test")
		if def.Name != "must-get-test" {
			t.Errorf("MustGet returned wrong provider")
		}
	})

	t.Run("not found panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for nonexistent provider")
			}
		}()
		MustGet("nonexistent")
	})
}

func TestAllFlags(t *testing.T) {
	resetRegistry()

	Register(&ProviderDefinition{
		Name: "provider1",
		Flags: []FlagDefinition{
			{Name: "flag-a", Description: "Flag A"},
			{Name: "flag-b", Description: "Flag B"},
		},
		Factory: mockFactory,
	})

	Register(&ProviderDefinition{
		Name: "provider2",
		Flags: []FlagDefinition{
			{Name: "flag-b", Description: "Flag B duplicate"}, // Should be deduplicated
			{Name: "flag-c", Description: "Flag C"},
		},
		Factory: mockFactory,
	})

	flags := AllFlags()

	if len(flags) != 3 {
		t.Errorf("AllFlags() returned %d flags, want 3", len(flags))
	}

	// Verify sorted order
	expectedFlags := []string{"flag-a", "flag-b", "flag-c"}
	for i, flag := range flags {
		if flag.Name != expectedFlags[i] {
			t.Errorf("AllFlags()[%d].Name = %q, want %q", i, flag.Name, expectedFlags[i])
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	resetRegistry()

	// Pre-register one provider
	Register(&ProviderDefinition{
		Name:    "concurrent-test",
		Aliases: []string{"ct"},
		Factory: mockFactory,
	})

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Get("concurrent-test")
			Get("ct")
			All()
			Names()
			Exists("concurrent-test")
			AllFlags()
		}()
	}

	wg.Wait()
}
