// Package core provides the fundamental types and interfaces for security policy
// evaluation, including assets, policies, providers, and the evaluation engine.
package core

// Asset represents the target we are scanning (e.g., a Host, a Repository, a Container).
// It replaces the loose "config" map and "asset" context map.
type Asset struct {
	ID     string            `json:"id,omitempty"`
	Name   string            `json:"name,omitempty"`   // User friendly name
	FQDN   string            `json:"fqdn,omitempty"`   // Fully Qualified Domain Name
	Type   string            `json:"type,omitempty"`   // e.g. "host", "container"
	Config map[string]string `json:"config,omitempty"` // Provider-specific configuration (e.g. file path)
}
