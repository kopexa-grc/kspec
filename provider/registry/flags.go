// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// RegisterFlags registers all flags from a provider definition to a cobra command.
func RegisterFlags(cmd *cobra.Command, def *ProviderDefinition) {
	for _, flag := range def.Flags {
		RegisterFlag(cmd, flag)
	}
}

// RegisterFlag registers a single flag to a cobra command.
func RegisterFlag(cmd *cobra.Command, flag FlagDefinition) {
	// Skip if flag already exists
	if cmd.Flags().Lookup(flag.Name) != nil {
		return
	}

	flagType := flag.Type
	if flagType == "" {
		flagType = FlagTypeString
	}

	switch flagType {
	case FlagTypeBool:
		defaultVal := flag.Default == "true"
		if flag.Short != "" {
			cmd.Flags().BoolP(flag.Name, flag.Short, defaultVal, flag.Description)
		} else {
			cmd.Flags().Bool(flag.Name, defaultVal, flag.Description)
		}
	case FlagTypeStringSlice:
		var defaultVal []string
		if flag.Default != "" {
			defaultVal = strings.Split(flag.Default, ",")
		}
		if flag.Short != "" {
			cmd.Flags().StringSliceP(flag.Name, flag.Short, defaultVal, flag.Description)
		} else {
			cmd.Flags().StringSlice(flag.Name, defaultVal, flag.Description)
		}
	case FlagTypeString:
		if flag.Short != "" {
			cmd.Flags().StringP(flag.Name, flag.Short, flag.Default, flag.Description)
		} else {
			cmd.Flags().String(flag.Name, flag.Default, flag.Description)
		}
	}
}

// RegisterAllProviderFlags registers flags from all providers to a cobra command.
func RegisterAllProviderFlags(cmd *cobra.Command) {
	for _, flag := range AllFlags() {
		RegisterFlag(cmd, flag)
	}
}

// BuildConfig builds a config map from cobra flags for a provider.
// It extracts flag values and applies environment variable fallbacks.
func BuildConfig(cmd *cobra.Command, def *ProviderDefinition) map[string]string {
	config := make(map[string]string)

	for _, flag := range def.Flags {
		configKey := def.GetConfigKey(flag.Name)
		value := getFlagValue(cmd, flag)

		if value != "" {
			config[configKey] = value
		}
	}

	return config
}

// getFlagValue gets the value of a flag, with environment variable fallback.
func getFlagValue(cmd *cobra.Command, flag FlagDefinition) string {
	flagType := flag.Type
	if flagType == "" {
		flagType = FlagTypeString
	}

	var value string

	switch flagType {
	case FlagTypeBool:
		if val, err := cmd.Flags().GetBool(flag.Name); err == nil && val {
			value = "true"
		}
	case FlagTypeStringSlice:
		if val, err := cmd.Flags().GetStringSlice(flag.Name); err == nil && len(val) > 0 {
			value = strings.Join(val, ",")
		}
	case FlagTypeString:
		val, _ := cmd.Flags().GetString(flag.Name) //nolint:errcheck // GetString returns empty string on error, which is acceptable
		value = val
	}

	// Apply environment variable fallback if value is empty
	if value == "" && flag.EnvVar != "" {
		value = os.Getenv(flag.EnvVar)
	}

	return value
}

// ParseAssetArgs parses positional arguments into an asset config.
// Returns the asset type name, asset name, and config map.
func ParseAssetArgs(def *ProviderDefinition, args []string) (assetType, assetName string, config map[string]string, err error) {
	config = make(map[string]string)

	if len(args) == 0 {
		// Use first asset type as default if only one exists
		if len(def.AssetTypes) == 1 {
			assetType = def.AssetTypes[0].Name
			assetName = def.AssetTypes[0].DefaultName
			if assetName == "" {
				assetName = "default"
			}
			return assetType, assetName, config, nil
		}
		return "", "", nil, nil
	}

	// First argument is the asset type
	assetType = args[0]
	assetDef, ok := def.GetAssetType(assetType)
	if !ok {
		// If only one asset type, treat first arg as asset name
		if len(def.AssetTypes) == 1 {
			assetType = def.AssetTypes[0].Name
			assetDef = &def.AssetTypes[0]
			assetName = args[0]
			args = args[1:]
		} else {
			return "", "", nil, nil
		}
	} else {
		args = args[1:]
	}

	// Parse positional arguments
	for i, argDef := range assetDef.Args {
		if i < len(args) {
			if argDef.ConfigKey != "" {
				config[argDef.ConfigKey] = args[i]
			}
			if i == 0 && assetName == "" {
				assetName = args[i]
			}
		}
	}

	// Use default name if not set
	if assetName == "" {
		assetName = assetDef.DefaultName
		if assetName == "" {
			assetName = "default"
		}
	}

	return assetType, assetName, config, nil
}
