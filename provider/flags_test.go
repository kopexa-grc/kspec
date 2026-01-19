// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package provider

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func newTestCommand() *cobra.Command {
	return &cobra.Command{
		Use: "test",
	}
}

func TestRegisterFlag(t *testing.T) {
	tests := []struct {
		name    string
		flag    FlagDefinition
		checkFn func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name: "string flag without short",
			flag: FlagDefinition{
				Name:        "my-flag",
				Description: "Test flag",
				Default:     "default-value",
			},
			checkFn: func(t *testing.T, cmd *cobra.Command) {
				f := cmd.Flags().Lookup("my-flag")
				if f == nil {
					t.Fatal("flag not registered")
				}
				if f.DefValue != "default-value" {
					t.Errorf("default = %q, want %q", f.DefValue, "default-value")
				}
				if f.Shorthand != "" {
					t.Errorf("shorthand = %q, want empty", f.Shorthand)
				}
			},
		},
		{
			name: "string flag with short",
			flag: FlagDefinition{
				Name:        "short-flag",
				Short:       "s",
				Description: "Short test flag",
			},
			checkFn: func(t *testing.T, cmd *cobra.Command) {
				f := cmd.Flags().Lookup("short-flag")
				if f == nil {
					t.Fatal("flag not registered")
				}
				if f.Shorthand != "s" {
					t.Errorf("shorthand = %q, want %q", f.Shorthand, "s")
				}
			},
		},
		{
			name: "bool flag",
			flag: FlagDefinition{
				Name:    "bool-flag",
				Type:    FlagTypeBool,
				Default: "true",
			},
			checkFn: func(t *testing.T, cmd *cobra.Command) {
				f := cmd.Flags().Lookup("bool-flag")
				if f == nil {
					t.Fatal("flag not registered")
				}
				if f.DefValue != "true" {
					t.Errorf("default = %q, want %q", f.DefValue, "true")
				}
			},
		},
		{
			name: "bool flag with short",
			flag: FlagDefinition{
				Name:  "bool-short",
				Short: "b",
				Type:  FlagTypeBool,
			},
			checkFn: func(t *testing.T, cmd *cobra.Command) {
				f := cmd.Flags().Lookup("bool-short")
				if f == nil {
					t.Fatal("flag not registered")
				}
				if f.Shorthand != "b" {
					t.Errorf("shorthand = %q, want %q", f.Shorthand, "b")
				}
			},
		},
		{
			name: "string slice flag",
			flag: FlagDefinition{
				Name:    "slice-flag",
				Type:    FlagTypeStringSlice,
				Default: "a,b,c",
			},
			checkFn: func(t *testing.T, cmd *cobra.Command) {
				f := cmd.Flags().Lookup("slice-flag")
				if f == nil {
					t.Fatal("flag not registered")
				}
				// Default for string slice is formatted as [a,b,c]
				if f.DefValue != "[a,b,c]" {
					t.Errorf("default = %q, want %q", f.DefValue, "[a,b,c]")
				}
			},
		},
		{
			name: "string slice flag with short",
			flag: FlagDefinition{
				Name:  "slice-short",
				Short: "l",
				Type:  FlagTypeStringSlice,
			},
			checkFn: func(t *testing.T, cmd *cobra.Command) {
				f := cmd.Flags().Lookup("slice-short")
				if f == nil {
					t.Fatal("flag not registered")
				}
				if f.Shorthand != "l" {
					t.Errorf("shorthand = %q, want %q", f.Shorthand, "l")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestCommand()
			RegisterFlag(cmd, tt.flag)
			tt.checkFn(t, cmd)
		})
	}
}

func TestRegisterFlag_SkipsDuplicate(t *testing.T) {
	cmd := newTestCommand()

	flag := FlagDefinition{
		Name:        "duplicate-flag",
		Description: "First registration",
		Default:     "first",
	}

	RegisterFlag(cmd, flag)

	// Try to register again with different values
	flag.Default = "second"
	flag.Description = "Second registration"
	RegisterFlag(cmd, flag)

	// Should keep first registration
	f := cmd.Flags().Lookup("duplicate-flag")
	if f == nil {
		t.Fatal("flag not found")
	}
	if f.DefValue != "first" {
		t.Errorf("default = %q, want %q (first registration)", f.DefValue, "first")
	}
}

func TestRegisterFlags(t *testing.T) {
	cmd := newTestCommand()
	def := &ProviderDefinition{
		Name: "test",
		Flags: []FlagDefinition{
			{Name: "flag-one", Default: "one"},
			{Name: "flag-two", Default: "two"},
			{Name: "flag-three", Type: FlagTypeBool},
		},
	}

	RegisterFlags(cmd, def)

	// Verify all flags are registered
	for _, flag := range def.Flags {
		if cmd.Flags().Lookup(flag.Name) == nil {
			t.Errorf("flag %q not registered", flag.Name)
		}
	}
}

func TestBuildConfig(t *testing.T) {
	tests := []struct {
		name     string
		def      *ProviderDefinition
		setup    func(cmd *cobra.Command)
		envSetup map[string]string
		want     map[string]string
	}{
		{
			name: "string flag value",
			def: &ProviderDefinition{
				Name: "test",
				Flags: []FlagDefinition{
					{Name: "my-flag"},
				},
			},
			setup: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set("my-flag", "test-value")
			},
			want: map[string]string{
				"my_flag": "test-value",
			},
		},
		{
			name: "config mapping applied",
			def: &ProviderDefinition{
				Name: "test",
				Flags: []FlagDefinition{
					{Name: "my-flag"},
				},
				ConfigMapping: map[string]string{
					"my-flag": "custom_key",
				},
			},
			setup: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set("my-flag", "value")
			},
			want: map[string]string{
				"custom_key": "value",
			},
		},
		{
			name: "empty value not included",
			def: &ProviderDefinition{
				Name: "test",
				Flags: []FlagDefinition{
					{Name: "empty-flag"},
					{Name: "set-flag"},
				},
			},
			setup: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set("set-flag", "value")
			},
			want: map[string]string{
				"set_flag": "value",
			},
		},
		{
			name: "env var fallback",
			def: &ProviderDefinition{
				Name: "test",
				Flags: []FlagDefinition{
					{Name: "env-flag", EnvVar: "TEST_ENV_VAR"},
				},
			},
			setup:    func(cmd *cobra.Command) {},
			envSetup: map[string]string{"TEST_ENV_VAR": "env-value"},
			want: map[string]string{
				"env_flag": "env-value",
			},
		},
		{
			name: "flag value takes precedence over env",
			def: &ProviderDefinition{
				Name: "test",
				Flags: []FlagDefinition{
					{Name: "priority-flag", EnvVar: "TEST_PRIORITY_VAR"},
				},
			},
			setup: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set("priority-flag", "flag-value")
			},
			envSetup: map[string]string{"TEST_PRIORITY_VAR": "env-value"},
			want: map[string]string{
				"priority_flag": "flag-value",
			},
		},
		{
			name: "bool flag true",
			def: &ProviderDefinition{
				Name: "test",
				Flags: []FlagDefinition{
					{Name: "bool-flag", Type: FlagTypeBool},
				},
			},
			setup: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set("bool-flag", "true")
			},
			want: map[string]string{
				"bool_flag": "true",
			},
		},
		{
			name: "bool flag false not included",
			def: &ProviderDefinition{
				Name: "test",
				Flags: []FlagDefinition{
					{Name: "bool-flag", Type: FlagTypeBool},
				},
			},
			setup: func(cmd *cobra.Command) {},
			want:  map[string]string{},
		},
		{
			name: "string slice flag",
			def: &ProviderDefinition{
				Name: "test",
				Flags: []FlagDefinition{
					{Name: "slice-flag", Type: FlagTypeStringSlice},
				},
			},
			setup: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set("slice-flag", "a,b,c")
			},
			want: map[string]string{
				"slice_flag": "a,b,c",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment and cleanup
			envKeys := make([]string, 0, len(tt.envSetup))
			for k, v := range tt.envSetup {
				os.Setenv(k, v)
				envKeys = append(envKeys, k)
			}
			t.Cleanup(func() {
				for _, k := range envKeys {
					os.Unsetenv(k)
				}
			})

			cmd := newTestCommand()
			RegisterFlags(cmd, tt.def)
			tt.setup(cmd)

			got := BuildConfig(cmd, tt.def)

			// Compare maps
			if len(got) != len(tt.want) {
				t.Errorf("config has %d entries, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				if gotV, ok := got[k]; !ok {
					t.Errorf("config missing key %q", k)
				} else if gotV != wantV {
					t.Errorf("config[%q] = %q, want %q", k, gotV, wantV)
				}
			}
		})
	}
}

func TestParseAssetArgs(t *testing.T) {
	singleAssetDef := &ProviderDefinition{
		Name: "single",
		AssetTypes: []AssetDefinition{
			{
				Name:        "account",
				DefaultName: "default-account",
				Args: []ArgDefinition{
					{Name: "id", Required: true, ConfigKey: "account_id"},
				},
			},
		},
	}

	multiAssetDef := &ProviderDefinition{
		Name: "multi",
		AssetTypes: []AssetDefinition{
			{
				Name: "org",
				Args: []ArgDefinition{
					{Name: "name", Required: true, ConfigKey: "org_name"},
				},
			},
			{
				Name:    "repo",
				Aliases: []string{"repository"},
				Args: []ArgDefinition{
					{Name: "path", Required: true, ConfigKey: "repo_path"},
				},
			},
		},
	}

	tests := []struct {
		name          string
		def           *ProviderDefinition
		args          []string
		wantAssetType string
		wantAssetName string
		wantConfig    map[string]string
	}{
		{
			name:          "empty args with single asset type",
			def:           singleAssetDef,
			args:          []string{},
			wantAssetType: "account",
			wantAssetName: "default-account",
			wantConfig:    map[string]string{},
		},
		{
			name:          "empty args with multi asset type",
			def:           multiAssetDef,
			args:          []string{},
			wantAssetType: "",
			wantAssetName: "",
			wantConfig:    nil,
		},
		{
			name:          "single asset type treats first arg as asset name",
			def:           singleAssetDef,
			args:          []string{"my-account"},
			wantAssetType: "account",
			wantAssetName: "my-account",
			wantConfig:    map[string]string{"account_id": "my-account"},
		},
		{
			name:          "multi asset type requires asset type arg",
			def:           multiAssetDef,
			args:          []string{"org", "my-org"},
			wantAssetType: "org",
			wantAssetName: "my-org",
			wantConfig:    map[string]string{"org_name": "my-org"},
		},
		{
			name:          "asset type alias works",
			def:           multiAssetDef,
			args:          []string{"repository", "my-repo"},
			wantAssetType: "repo",
			wantAssetName: "my-repo",
			wantConfig:    map[string]string{"repo_path": "my-repo"},
		},
		{
			name:          "unknown asset type with multi returns empty",
			def:           multiAssetDef,
			args:          []string{"unknown"},
			wantAssetType: "",
			wantAssetName: "",
			wantConfig:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assetType, assetName, config, err := ParseAssetArgs(tt.def, tt.args)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if assetType != tt.wantAssetType {
				t.Errorf("assetType = %q, want %q", assetType, tt.wantAssetType)
			}

			if assetName != tt.wantAssetName {
				t.Errorf("assetName = %q, want %q", assetName, tt.wantAssetName)
			}

			if tt.wantConfig == nil {
				if config != nil {
					t.Errorf("config = %v, want nil", config)
				}
			} else {
				if len(config) != len(tt.wantConfig) {
					t.Errorf("config has %d entries, want %d", len(config), len(tt.wantConfig))
				}
				for k, wantV := range tt.wantConfig {
					if gotV, ok := config[k]; !ok {
						t.Errorf("config missing key %q", k)
					} else if gotV != wantV {
						t.Errorf("config[%q] = %q, want %q", k, gotV, wantV)
					}
				}
			}
		})
	}
}

func TestParseAssetArgs_DefaultName(t *testing.T) {
	// Test that default name is used when no args provided
	def := &ProviderDefinition{
		Name: "test",
		AssetTypes: []AssetDefinition{
			{
				Name:        "resource",
				DefaultName: "my-default",
			},
		},
	}

	assetType, assetName, config, err := ParseAssetArgs(def, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if assetType != "resource" {
		t.Errorf("assetType = %q, want %q", assetType, "resource")
	}
	if config == nil {
		t.Errorf("config is nil, want empty map")
	}
	if assetName != "my-default" {
		t.Errorf("assetName = %q, want %q", assetName, "my-default")
	}

	// Test fallback to "default" when no DefaultName
	defNoDefault := &ProviderDefinition{
		Name: "test",
		AssetTypes: []AssetDefinition{
			{Name: "resource"},
		},
	}

	assetType, assetName, config, err = ParseAssetArgs(defNoDefault, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if assetType != "resource" {
		t.Errorf("assetType = %q, want %q", assetType, "resource")
	}
	if config == nil {
		t.Errorf("config is nil, want empty map")
	}
	if assetName != "default" {
		t.Errorf("assetName = %q, want %q", assetName, "default")
	}
}
