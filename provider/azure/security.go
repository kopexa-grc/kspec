// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"

	"github.com/kopexa-grc/kspec/core"
)

// SecurityAssessmentResource represents an Azure Security Center assessment resource scanner.
type SecurityAssessmentResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *SecurityAssessmentResource) Name() string {
	return "azure_security_assessment"
}

// Fetch retrieves security assessments from Azure Security Center.
func (r *SecurityAssessmentResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	scope := fmt.Sprintf("/subscriptions/%s", r.subscriptionID)
	client, err := armsecurity.NewAssessmentsClient(r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create security assessments client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(scope, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list security assessments: %w", err)
		}

		for _, assessment := range page.Value {
			data, err := json.Marshal(assessment)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			if assessment.Properties != nil {
				resourceMap["displayName"] = assessment.Properties.DisplayName

				if assessment.Properties.Status != nil {
					resourceMap["statusCode"] = assessment.Properties.Status.Code
					resourceMap["statusCause"] = assessment.Properties.Status.Cause
					resourceMap["statusDescription"] = assessment.Properties.Status.Description
				}

				if assessment.Properties.ResourceDetails != nil {
					resourceMap["resourceSource"] = assessment.Properties.ResourceDetails.GetResourceDetails().Source
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}

// SecurityAlertResource represents an Azure Defender security alert resource scanner.
type SecurityAlertResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *SecurityAlertResource) Name() string {
	return "azure_security_alert"
}

// Fetch retrieves security alerts from Microsoft Defender for Cloud.
func (r *SecurityAlertResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armsecurity.NewAlertsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create security alerts client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list security alerts: %w", err)
		}

		for _, alert := range page.Value {
			data, err := json.Marshal(alert)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			if alert.Properties != nil {
				resourceMap["alertDisplayName"] = alert.Properties.AlertDisplayName
				resourceMap["alertType"] = alert.Properties.AlertType
				resourceMap["severity"] = alert.Properties.Severity
				resourceMap["status"] = alert.Properties.Status
				resourceMap["compromisedEntity"] = alert.Properties.CompromisedEntity
				resourceMap["intent"] = alert.Properties.Intent
				resourceMap["startTimeUtc"] = alert.Properties.StartTimeUTC
				resourceMap["endTimeUtc"] = alert.Properties.EndTimeUTC
				resourceMap["isIncident"] = alert.Properties.IsIncident
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}

// DefenderSettingResource represents Microsoft Defender for Cloud settings resource scanner.
type DefenderSettingResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *DefenderSettingResource) Name() string {
	return "azure_defender_setting"
}

// Fetch retrieves Defender for Cloud settings.
func (r *DefenderSettingResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armsecurity.NewPricingsClient(r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Defender pricing client: %w", err)
	}

	resources := make([]core.Resource, 0)

	scope := fmt.Sprintf("subscriptions/%s", r.subscriptionID)
	result, err := client.List(ctx, scope, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list Defender pricings: %w", err)
	}

	for _, pricing := range result.Value {
		data, err := json.Marshal(pricing)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		if pricing.Properties != nil {
			resourceMap["pricingTier"] = pricing.Properties.PricingTier
			resourceMap["subPlan"] = pricing.Properties.SubPlan
			resourceMap["freeTrialRemainingTime"] = pricing.Properties.FreeTrialRemainingTime
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

// SecureScoreResource represents an Azure Secure Score resource scanner.
type SecureScoreResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *SecureScoreResource) Name() string {
	return "azure_secure_score"
}

// Fetch retrieves secure scores from Azure Security Center.
func (r *SecureScoreResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armsecurity.NewSecureScoresClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure scores client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list secure scores: %w", err)
		}

		for _, score := range page.Value {
			data, err := json.Marshal(score)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			if score.Properties != nil {
				resourceMap["displayName"] = score.Properties.DisplayName
				resourceMap["currentScore"] = score.Properties.Score.Current
				resourceMap["maxScore"] = score.Properties.Score.Max
				resourceMap["percentage"] = score.Properties.Score.Percentage
				resourceMap["weight"] = score.Properties.Weight
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
