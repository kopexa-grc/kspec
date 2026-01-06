package scanner

import (
	"fmt"

	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/core"
)

// ScanConfig holds the configuration for a scan
type ScanConfig struct {
	// Provider configuration
	ProviderName   string
	ProviderConfig map[string]string

	// Asset configuration
	Asset core.Asset

	// Policies to evaluate
	Policies []core.Policy
}

// ScanEventType represents the type of scan event
type ScanEventType int

const (
	// EventTreeCreated is sent when the initial tree is created
	EventTreeCreated ScanEventType = iota
	// EventTreeUpdated is sent when the tree is updated
	EventTreeUpdated
	// EventDiscoveryStarted is sent when discovery begins
	EventDiscoveryStarted
	// EventDiscoveryComplete is sent when discovery completes
	EventDiscoveryComplete
	// EventScanStarted is sent when scanning begins
	EventScanStarted
	// EventScanComplete is sent when scanning completes
	EventScanComplete
	// EventResourceScanning is sent when a resource type starts scanning
	EventResourceScanning
	// EventResourceComplete is sent when a resource type completes
	EventResourceComplete
	// EventError is sent when an error occurs
	EventError
)

// ScanEvent represents an event during scanning
type ScanEvent struct {
	Type         ScanEventType
	Tree         *common.ResourceTree
	ResourceType string
	Error        error
	Message      string
}

// ScanError represents an error that occurred during scanning
type ScanError struct {
	Phase        string // "discovery", "fetch", "evaluation"
	ResourceType string
	Message      string
	Err          error
}

// Error implements the error interface.
func (e *ScanError) Error() string {
	if e.ResourceType != "" {
		return fmt.Sprintf("%s error for %s: %s", e.Phase, e.ResourceType, e.Message)
	}
	return fmt.Sprintf("%s error: %s", e.Phase, e.Message)
}

// Unwrap returns the underlying error.
func (e *ScanError) Unwrap() error {
	return e.Err
}

// ScanResult contains the results of a scan including any errors
type ScanResult struct {
	Tree   *common.ResourceTree
	Errors []*ScanError
}

// ScanEventHandler is a callback for scan events
type ScanEventHandler func(event ScanEvent)

// CheckResult wraps common.CheckResult for the scanner package
type CheckResult = common.CheckResult

// ResourceOrder defines the preferred order for resource types per provider
var ResourceOrder = map[string][]string{
	"github": {"github_organization", "github_team", "github_repo"},
	"azure": {
		"azure_subscription", "azure_storage_account", "azure_sql_server", "azure_sql_database",
		"azure_mysql_server", "azure_postgresql_server", "azure_keyvault_vault",
		"azure_network_security_group", "azure_virtual_machine", "azure_app_service",
	},
	"ms365": {
		"ms365_tenant", "ms365_user", "ms365_group", "ms365_group_lifecycle_policy",
		"ms365_application", "ms365_service_principal",
		"ms365_device", "ms365_managed_device", "ms365_device_configuration", "ms365_device_compliance_policy",
		"ms365_domain", "ms365_conditional_access_policy", "ms365_named_location",
		"ms365_directory_role", "ms365_role_assignment", "ms365_authorization_policy",
		"ms365_authentication_methods_policy", "ms365_security_defaults_policy",
		"ms365_risky_user", "ms365_secure_score", "ms365_security_alert",
		"ms365_team", "ms365_teams_settings", "ms365_directory_setting", "ms365_external_identity_policy",
	},
}
