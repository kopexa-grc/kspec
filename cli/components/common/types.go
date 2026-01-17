package common

// AssetState represents the current state of an asset
type AssetState string

// AssetState constants define the possible states of an asset during scanning.
const (
	AssetStatePending   AssetState = "pending"
	AssetStateDiscovery AssetState = "discovering"
	AssetStateScanning  AssetState = "scanning"
	AssetStateComplete  AssetState = "complete"
	AssetStateError     AssetState = "error"
)

// Asset represents a scan target with its current state and results
type Asset struct {
	Type          string
	Name          string
	ID            string
	State         AssetState
	ResourceCount int
	ChecksPassed  int
	ChecksFailed  int
	ChecksSkipped int
	Error         error
}

// ViewMode represents the current UI view
type ViewMode string

// ViewMode constants define the possible view modes in the CLI.
const (
	ViewModeDiscovery ViewMode = "discovery"
	ViewModeOverview  ViewMode = "overview"
	ViewModeDetail    ViewMode = "detail"
)

// CheckResult represents a single check result
type CheckResult struct {
	ID          string
	Group       string
	Name        string
	Status      string
	Details     string
	Severity    string
	Remediation string // Markdown remediation guidance
	Docs        string // Markdown documentation
	Audit       string // Markdown audit information
}
