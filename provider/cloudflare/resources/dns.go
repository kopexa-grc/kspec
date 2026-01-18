// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"

	"github.com/kopexa-grc/kspec/core"
)

// DNS record type constants
const (
	RecordTypeTXT = "TXT"
	RecordTypeMX  = "MX"
)

// DNSRecord fetches DNS records for a zone.
type DNSRecord struct {
	client *cloudflare.Client
}

// NewDNSRecord creates a new DNSRecord resource.
func NewDNSRecord(client *cloudflare.Client) *DNSRecord {
	return &DNSRecord{client: client}
}

// Name returns the resource type name.
func (r *DNSRecord) Name() string {
	return "cloudflare_dns_record"
}

// Fetch retrieves all DNS records for a zone.
func (r *DNSRecord) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	zoneID := asset.Config["zone_id"]
	if zoneID == "" {
		return nil, nil
	}

	var resources []core.Resource

	iter := r.client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
	})

	for iter.Next() {
		record := iter.Current()

		data, err := json.Marshal(record)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		// Add security-relevant computed fields
		r.addSecurityFields(resourceMap)

		resources = append(resources, resourceMap)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *DNSRecord) addSecurityFields(record map[string]interface{}) {
	recordType := ""
	if rt, ok := record["type"].(string); ok {
		recordType = rt
	}
	content := ""
	if c, ok := record["content"].(string); ok {
		content = c
	}
	name := ""
	if n, ok := record["name"].(string); ok {
		name = n
	}

	// Check for email security records
	record["is_spf"] = recordType == RecordTypeTXT && strings.HasPrefix(content, "v=spf1")
	record["is_dkim"] = recordType == RecordTypeTXT && strings.Contains(name, "._domainkey.")
	record["is_dmarc"] = recordType == RecordTypeTXT && strings.HasPrefix(name, "_dmarc.")

	// Check for MX records
	record["is_mx"] = recordType == RecordTypeMX

	// Check for wildcard records
	record["is_wildcard"] = strings.HasPrefix(name, "*.")

	// Check if proxied (orange cloud)
	if proxied, ok := record["proxied"].(bool); ok {
		record["is_proxied"] = proxied
	}

	// TTL check
	if ttl, ok := record["ttl"].(float64); ok {
		record["ttl_seconds"] = int(ttl)
		record["ttl_auto"] = ttl == 1
	}
}
