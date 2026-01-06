package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// FirewallResource fetches Hetzner Cloud firewalls.
type FirewallResource struct {
	client *hcloud.Client
}

// Name returns the resource type name.
func (r *FirewallResource) Name() string {
	return "hcloud_firewall"
}

// Fetch retrieves all Hetzner Cloud firewalls.
func (r *FirewallResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	firewalls, err := r.client.Firewall.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(firewalls))

	for _, fw := range firewalls {
		resource := make(core.Resource)
		resource["id"] = fw.ID
		resource["name"] = fw.Name
		resource["created"] = fw.Created.String()
		resource["labels"] = fw.Labels

		// Rules
		var rules []map[string]interface{}
		var inboundRuleCount, outboundRuleCount int
		var allowsAllInbound, allowsAllOutbound bool

		for _, rule := range fw.Rules {
			ruleMap := map[string]interface{}{
				"direction":   string(rule.Direction),
				"protocol":    string(rule.Protocol),
				"description": rule.Description,
			}

			if rule.Port != nil {
				ruleMap["port"] = *rule.Port
			}

			// Source/destination IPs
			var sourceIPs, destIPs []string
			for _, ip := range rule.SourceIPs {
				sourceIPs = append(sourceIPs, ip.String())
			}
			for _, ip := range rule.DestinationIPs {
				destIPs = append(destIPs, ip.String())
			}
			ruleMap["source_ips"] = sourceIPs
			ruleMap["destination_ips"] = destIPs

			rules = append(rules, ruleMap)

			// Count by direction
			switch rule.Direction {
			case hcloud.FirewallRuleDirectionIn:
				inboundRuleCount++
				// Check for overly permissive rules (0.0.0.0/0 or ::/0)
				for _, ip := range rule.SourceIPs {
					if ip.String() == "0.0.0.0/0" || ip.String() == "::/0" {
						allowsAllInbound = true
					}
				}
			case hcloud.FirewallRuleDirectionOut:
				outboundRuleCount++
				for _, ip := range rule.DestinationIPs {
					if ip.String() == "0.0.0.0/0" || ip.String() == "::/0" {
						allowsAllOutbound = true
					}
				}
			}
		}
		resource["rules"] = rules
		resource["rule_count"] = len(rules)
		resource["inbound_rule_count"] = inboundRuleCount
		resource["outbound_rule_count"] = outboundRuleCount

		// Applied to resources
		var appliedServers, appliedLabelSelectors []interface{}
		for _, applied := range fw.AppliedTo {
			switch applied.Type {
			case hcloud.FirewallResourceTypeServer:
				if applied.Server != nil {
					appliedServers = append(appliedServers, applied.Server.ID)
				}
			case hcloud.FirewallResourceTypeLabelSelector:
				if applied.LabelSelector != nil {
					appliedLabelSelectors = append(appliedLabelSelectors, applied.LabelSelector.Selector)
				}
			}
		}
		resource["applied_to_servers"] = appliedServers
		resource["applied_to_label_selectors"] = appliedLabelSelectors
		resource["applied_to_server_count"] = len(appliedServers)
		resource["applied_to_label_selector_count"] = len(appliedLabelSelectors)

		// Security flags
		resource["has_rules"] = len(fw.Rules) > 0
		resource["has_inbound_rules"] = inboundRuleCount > 0
		resource["has_outbound_rules"] = outboundRuleCount > 0
		resource["is_applied"] = len(fw.AppliedTo) > 0
		resource["allows_all_inbound"] = allowsAllInbound
		resource["allows_all_outbound"] = allowsAllOutbound
		resource["has_labels"] = len(fw.Labels) > 0

		resources = append(resources, resource)
	}

	return resources, nil
}
