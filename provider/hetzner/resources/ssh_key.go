// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"strings"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// SSH key type constants
const (
	keyTypeRSA     = "rsa"
	keyTypeED25519 = "ed25519"
	keyTypeECDSA   = "ecdsa"
	keyTypeDSA     = "dsa"
	keyTypeUnknown = "unknown"
)

// SSHKey fetches Hetzner Cloud SSH keys.
type SSHKey struct {
	client *hcloud.Client
}

// NewSSHKey creates a new SSHKey resource.
func NewSSHKey(client *hcloud.Client) *SSHKey {
	return &SSHKey{client: client}
}

// Name returns the resource type name.
func (r *SSHKey) Name() string {
	return "hcloud_ssh_key"
}

// Fetch retrieves all Hetzner Cloud SSH keys.
func (r *SSHKey) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	sshKeys, err := r.client.SSHKey.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(sshKeys))

	for _, key := range sshKeys {
		resource := make(core.Resource)
		resource["id"] = key.ID
		resource["name"] = key.Name
		resource["created"] = key.Created.String()
		resource["labels"] = key.Labels
		resource["fingerprint"] = key.Fingerprint

		// Extract key type from public key
		keyType := detectSSHKeyType(key.PublicKey)
		resource["key_type"] = keyType

		// Store public key (but not the full content for security)
		resource["public_key_prefix"] = truncateKey(key.PublicKey, 50)

		// Security flags
		resource["is_rsa"] = keyType == keyTypeRSA
		resource["is_ed25519"] = keyType == keyTypeED25519
		resource["is_ecdsa"] = keyType == keyTypeECDSA
		resource["is_dsa"] = keyType == keyTypeDSA
		resource["is_weak_key_type"] = keyType == keyTypeDSA || keyType == keyTypeRSA // DSA is deprecated, RSA is weaker than ed25519
		resource["has_labels"] = len(key.Labels) > 0

		resources = append(resources, resource)
	}

	return resources, nil
}

// detectSSHKeyType detects the SSH key type from the public key prefix.
func detectSSHKeyType(publicKey string) string {
	switch {
	case strings.HasPrefix(publicKey, "ssh-rsa"):
		return keyTypeRSA
	case strings.HasPrefix(publicKey, "ssh-ed25519"):
		return keyTypeED25519
	case strings.HasPrefix(publicKey, "ecdsa-"):
		return keyTypeECDSA
	case strings.HasPrefix(publicKey, "ssh-dss"):
		return keyTypeDSA
	default:
		return keyTypeUnknown
	}
}

// truncateKey truncates a public key for display purposes.
func truncateKey(key string, maxLen int) string {
	if len(key) <= maxLen {
		return key
	}
	return key[:maxLen] + "..."
}
