package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"

	"github.com/kopexa-grc/kspec/core"
)

// WAFWebACLResource fetches WAFv2 Web ACLs.
type WAFWebACLResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *WAFWebACLResource) Name() string {
	return "aws_waf_webacl"
}

// Fetch retrieves all WAFv2 Web ACLs across configured regions.
func (r *WAFWebACLResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// WAFv2 has two scopes: REGIONAL and CLOUDFRONT
	// CLOUDFRONT scope is only available in us-east-1
	scopes := []struct {
		scope   types.Scope
		regions []string
	}{
		{types.ScopeRegional, r.conn.regions},
		{types.ScopeCloudfront, []string{"us-east-1"}},
	}

	for _, scopeConfig := range scopes {
		for _, region := range scopeConfig.regions {
			client := wafv2.NewFromConfig(r.conn.ConfigForRegion(region))

			var nextMarker *string
			for {
				resp, err := client.ListWebACLs(ctx, &wafv2.ListWebACLsInput{
					Scope:      scopeConfig.scope,
					NextMarker: nextMarker,
					Limit:      aws.Int32(100),
				})
				if err != nil {
					break
				}

				for _, aclSummary := range resp.WebACLs {
					// Get full ACL details
					aclResp, err := client.GetWebACL(ctx, &wafv2.GetWebACLInput{
						Name:  aclSummary.Name,
						Scope: scopeConfig.scope,
						Id:    aclSummary.Id,
					})
					if err != nil {
						continue
					}

					acl := aclResp.WebACL
					resource := make(core.Resource)
					resource["id"] = aws.ToString(acl.Id)
					resource["arn"] = aws.ToString(acl.ARN)
					resource["name"] = aws.ToString(acl.Name)
					resource["region"] = region
					resource["scope"] = string(scopeConfig.scope)

					resource["description"] = aws.ToString(acl.Description)
					resource["capacity"] = acl.Capacity

					// Default action
					if acl.DefaultAction != nil {
						if acl.DefaultAction.Allow != nil {
							resource["default_action"] = "ALLOW"
						} else if acl.DefaultAction.Block != nil {
							resource["default_action"] = "BLOCK"
						}
					}

					// Visibility config
					if acl.VisibilityConfig != nil {
						resource["cloudwatch_metrics_enabled"] = acl.VisibilityConfig.CloudWatchMetricsEnabled
						resource["metric_name"] = aws.ToString(acl.VisibilityConfig.MetricName)
						resource["sampled_requests_enabled"] = acl.VisibilityConfig.SampledRequestsEnabled
					}

					// Rules
					resource["rule_count"] = len(acl.Rules)
					ruleNames := make([]string, 0, len(acl.Rules))
					var managedRuleGroups []string
					hasRateLimiting := false
					hasSQLi := false
					hasXSS := false

					for _, rule := range acl.Rules {
						ruleNames = append(ruleNames, aws.ToString(rule.Name))

						// Check for managed rule groups
						if rule.Statement != nil && rule.Statement.ManagedRuleGroupStatement != nil {
							mrg := rule.Statement.ManagedRuleGroupStatement
							managedRuleGroups = append(managedRuleGroups, aws.ToString(mrg.Name))

							// Check for common protections
							name := aws.ToString(mrg.Name)
							if name == "AWSManagedRulesCommonRuleSet" {
								hasXSS = true
								hasSQLi = true
							}
							if name == "AWSManagedRulesSQLiRuleSet" {
								hasSQLi = true
							}
							if name == "AWSManagedRulesKnownBadInputsRuleSet" {
								hasXSS = true
							}
						}

						// Check for rate-based rules
						if rule.Statement != nil && rule.Statement.RateBasedStatement != nil {
							hasRateLimiting = true
						}
					}

					resource["rule_names"] = ruleNames
					resource["managed_rule_groups"] = managedRuleGroups
					resource["managed_rule_count"] = len(managedRuleGroups)
					resource["has_rate_limiting"] = hasRateLimiting
					resource["has_sqli_protection"] = hasSQLi
					resource["has_xss_protection"] = hasXSS

					// Get associated resources
					assocResp, err := client.ListResourcesForWebACL(ctx, &wafv2.ListResourcesForWebACLInput{
						WebACLArn: acl.ARN,
					})
					if err == nil {
						resource["associated_resources"] = assocResp.ResourceArns
						resource["associated_resource_count"] = len(assocResp.ResourceArns)
						resource["is_attached"] = len(assocResp.ResourceArns) > 0
					} else {
						resource["is_attached"] = false
						resource["associated_resource_count"] = 0
					}

					// Get logging configuration
					loggingResp, err := client.GetLoggingConfiguration(ctx, &wafv2.GetLoggingConfigurationInput{
						ResourceArn: acl.ARN,
					})
					if err == nil && loggingResp.LoggingConfiguration != nil {
						resource["has_logging"] = true
						resource["log_destinations"] = loggingResp.LoggingConfiguration.LogDestinationConfigs

						// Redacted fields
						var redactedFields []string
						for _, rf := range loggingResp.LoggingConfiguration.RedactedFields {
							if rf.SingleHeader != nil {
								redactedFields = append(redactedFields, "Header:"+aws.ToString(rf.SingleHeader.Name))
							}
							if rf.QueryString != nil {
								redactedFields = append(redactedFields, "QueryString")
							}
							if rf.Body != nil {
								redactedFields = append(redactedFields, "Body")
							}
						}
						resource["redacted_fields"] = redactedFields
					} else {
						resource["has_logging"] = false
					}

					// Computed
					resource["is_regional"] = scopeConfig.scope == types.ScopeRegional
					resource["is_cloudfront"] = scopeConfig.scope == types.ScopeCloudfront
					resource["has_rules"] = len(acl.Rules) > 0

					resources = append(resources, resource)
				}

				if resp.NextMarker == nil {
					break
				}
				nextMarker = resp.NextMarker
			}
		}
	}

	return resources, nil
}

// WAFIPSetResource fetches WAFv2 IP Sets.
type WAFIPSetResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *WAFIPSetResource) Name() string {
	return "aws_waf_ipset"
}

// Fetch retrieves all WAFv2 IP Sets across configured regions.
func (r *WAFIPSetResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	scopes := []struct {
		scope   types.Scope
		regions []string
	}{
		{types.ScopeRegional, r.conn.regions},
		{types.ScopeCloudfront, []string{"us-east-1"}},
	}

	for _, scopeConfig := range scopes {
		for _, region := range scopeConfig.regions {
			client := wafv2.NewFromConfig(r.conn.ConfigForRegion(region))

			var nextMarker *string
			for {
				resp, err := client.ListIPSets(ctx, &wafv2.ListIPSetsInput{
					Scope:      scopeConfig.scope,
					NextMarker: nextMarker,
					Limit:      aws.Int32(100),
				})
				if err != nil {
					break
				}

				for _, ipSetSummary := range resp.IPSets {
					// Get full IP set details
					ipSetResp, err := client.GetIPSet(ctx, &wafv2.GetIPSetInput{
						Name:  ipSetSummary.Name,
						Scope: scopeConfig.scope,
						Id:    ipSetSummary.Id,
					})
					if err != nil {
						continue
					}

					ipSet := ipSetResp.IPSet
					resource := make(core.Resource)
					resource["id"] = aws.ToString(ipSet.Id)
					resource["arn"] = aws.ToString(ipSet.ARN)
					resource["name"] = aws.ToString(ipSet.Name)
					resource["region"] = region
					resource["scope"] = string(scopeConfig.scope)

					resource["description"] = aws.ToString(ipSet.Description)
					resource["ip_address_version"] = string(ipSet.IPAddressVersion)
					resource["addresses"] = ipSet.Addresses
					resource["address_count"] = len(ipSet.Addresses)

					// Computed
					resource["is_ipv4"] = ipSet.IPAddressVersion == types.IPAddressVersionIpv4
					resource["is_ipv6"] = ipSet.IPAddressVersion == types.IPAddressVersionIpv6
					resource["is_regional"] = scopeConfig.scope == types.ScopeRegional
					resource["is_cloudfront"] = scopeConfig.scope == types.ScopeCloudfront
					resource["has_addresses"] = len(ipSet.Addresses) > 0

					resources = append(resources, resource)
				}

				if resp.NextMarker == nil {
					break
				}
				nextMarker = resp.NextMarker
			}
		}
	}

	return resources, nil
}

// WAFRuleGroupResource fetches WAFv2 Rule Groups.
type WAFRuleGroupResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *WAFRuleGroupResource) Name() string {
	return "aws_waf_rulegroup"
}

// Fetch retrieves all WAFv2 Rule Groups across configured regions.
func (r *WAFRuleGroupResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	scopes := []struct {
		scope   types.Scope
		regions []string
	}{
		{types.ScopeRegional, r.conn.regions},
		{types.ScopeCloudfront, []string{"us-east-1"}},
	}

	for _, scopeConfig := range scopes {
		for _, region := range scopeConfig.regions {
			client := wafv2.NewFromConfig(r.conn.ConfigForRegion(region))

			var nextMarker *string
			for {
				resp, err := client.ListRuleGroups(ctx, &wafv2.ListRuleGroupsInput{
					Scope:      scopeConfig.scope,
					NextMarker: nextMarker,
					Limit:      aws.Int32(100),
				})
				if err != nil {
					break
				}

				for _, rgSummary := range resp.RuleGroups {
					// Get full rule group details
					rgResp, err := client.GetRuleGroup(ctx, &wafv2.GetRuleGroupInput{
						Name:  rgSummary.Name,
						Scope: scopeConfig.scope,
						Id:    rgSummary.Id,
					})
					if err != nil {
						continue
					}

					rg := rgResp.RuleGroup
					resource := make(core.Resource)
					resource["id"] = aws.ToString(rg.Id)
					resource["arn"] = aws.ToString(rg.ARN)
					resource["name"] = aws.ToString(rg.Name)
					resource["region"] = region
					resource["scope"] = string(scopeConfig.scope)

					resource["description"] = aws.ToString(rg.Description)
					resource["capacity"] = rg.Capacity
					resource["rule_count"] = len(rg.Rules)

					// Visibility config
					if rg.VisibilityConfig != nil {
						resource["cloudwatch_metrics_enabled"] = rg.VisibilityConfig.CloudWatchMetricsEnabled
						resource["metric_name"] = aws.ToString(rg.VisibilityConfig.MetricName)
						resource["sampled_requests_enabled"] = rg.VisibilityConfig.SampledRequestsEnabled
					}

					// Rules
					ruleNames := make([]string, 0, len(rg.Rules))
					for _, rule := range rg.Rules {
						ruleNames = append(ruleNames, aws.ToString(rule.Name))
					}
					resource["rule_names"] = ruleNames

					// Computed
					resource["is_regional"] = scopeConfig.scope == types.ScopeRegional
					resource["is_cloudfront"] = scopeConfig.scope == types.ScopeCloudfront
					resource["has_rules"] = len(rg.Rules) > 0

					resources = append(resources, resource)
				}

				if resp.NextMarker == nil {
					break
				}
				nextMarker = resp.NextMarker
			}
		}
	}

	return resources, nil
}
