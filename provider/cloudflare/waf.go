package cloudflare

import (
	"context"
	"encoding/json"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/rulesets"

	"github.com/kopexa-grc/kspec/core"
)

// WAFRuleResource fetches WAF rulesets for a zone.
type WAFRuleResource struct {
	client *cloudflare.Client
}

// Name returns the resource type name.
func (r *WAFRuleResource) Name() string {
	return "cloudflare_waf_rule"
}

// Fetch retrieves WAF rulesets for a zone.
func (r *WAFRuleResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	zoneID := asset.Config["zone_id"]
	if zoneID == "" {
		return nil, nil
	}

	var resources []core.Resource

	// List zone rulesets
	result, err := r.client.Rulesets.List(ctx, rulesets.RulesetListParams{
		ZoneID: cloudflare.F(zoneID),
	})
	if err != nil {
		return nil, err
	}

	for _, ruleset := range result.Result {
		data, err := json.Marshal(ruleset)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		// Add security flags
		r.addSecurityFlags(resourceMap)

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

func (r *WAFRuleResource) addSecurityFlags(ruleset map[string]interface{}) {
	kind, _ := ruleset["kind"].(string)
	phase, _ := ruleset["phase"].(string)

	// Identify WAF managed rulesets
	ruleset["is_managed"] = kind == "managed"
	ruleset["is_custom"] = kind == "custom" || kind == "zone"

	// Identify OWASP ruleset
	if name, ok := ruleset["name"].(string); ok {
		ruleset["is_owasp"] = name == "Cloudflare OWASP Core Ruleset" ||
			name == "Cloudflare Managed Ruleset"
	}

	// Check phase for WAF rules
	ruleset["is_http_request_firewall"] = phase == "http_request_firewall_managed" ||
		phase == "http_request_firewall_custom"
}

// FirewallRuleResource fetches custom firewall rules for a zone.
type FirewallRuleResource struct {
	client *cloudflare.Client
}

// Name returns the resource type name.
func (r *FirewallRuleResource) Name() string {
	return "cloudflare_firewall_rule"
}

// Fetch retrieves custom firewall rules for a zone.
func (r *FirewallRuleResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	zoneID := asset.Config["zone_id"]
	if zoneID == "" {
		return nil, nil
	}

	var resources []core.Resource

	// Get custom rulesets for http_request_firewall_custom phase
	result, err := r.client.Rulesets.List(ctx, rulesets.RulesetListParams{
		ZoneID: cloudflare.F(zoneID),
	})
	if err != nil {
		return nil, err
	}

	for _, ruleset := range result.Result {
		// Only process custom firewall rules
		if ruleset.Phase != "http_request_firewall_custom" {
			continue
		}

		// Get full ruleset with rules
		fullRuleset, err := r.client.Rulesets.Get(ctx, ruleset.ID, rulesets.RulesetGetParams{
			ZoneID: cloudflare.F(zoneID),
		})
		if err != nil {
			continue
		}

		for _, rule := range fullRuleset.Rules {
			data, err := json.Marshal(rule)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			resourceMap["zone_id"] = zoneID
			resourceMap["ruleset_id"] = ruleset.ID

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
