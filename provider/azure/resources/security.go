// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"

	"github.com/kopexa-grc/kspec/core"
)

// SecurityAssessment represents an Azure Security Center assessment resource scanner.
type SecurityAssessment struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewSecurityAssessment creates a new SecurityAssessment resource scanner.
func NewSecurityAssessment(credential azcore.TokenCredential, subscriptionID string) *SecurityAssessment {
	return &SecurityAssessment{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *SecurityAssessment) Name() string {
	return "azure_security_assessment"
}

// Fetch retrieves security assessments from Azure Security Center.
func (r *SecurityAssessment) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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

// SecurityAlert represents an Azure Defender security alert resource scanner.
type SecurityAlert struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewSecurityAlert creates a new SecurityAlert resource scanner.
func NewSecurityAlert(credential azcore.TokenCredential, subscriptionID string) *SecurityAlert {
	return &SecurityAlert{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *SecurityAlert) Name() string {
	return "azure_security_alert"
}

// Fetch retrieves security alerts from Microsoft Defender for Cloud.
func (r *SecurityAlert) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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

// DefenderSetting represents Microsoft Defender for Cloud settings resource scanner.
type DefenderSetting struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewDefenderSetting creates a new DefenderSetting resource scanner.
func NewDefenderSetting(credential azcore.TokenCredential, subscriptionID string) *DefenderSetting {
	return &DefenderSetting{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *DefenderSetting) Name() string {
	return "azure_defender_setting"
}

// Fetch retrieves Defender for Cloud settings.
func (r *DefenderSetting) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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

// SecureScore represents an Azure Secure Score resource scanner.
type SecureScore struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewSecureScore creates a new SecureScore resource scanner.
func NewSecureScore(credential azcore.TokenCredential, subscriptionID string) *SecureScore {
	return &SecureScore{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *SecureScore) Name() string {
	return "azure_secure_score"
}

// Fetch retrieves secure scores from Azure Security Center.
func (r *SecureScore) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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
