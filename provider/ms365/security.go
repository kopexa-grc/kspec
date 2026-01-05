package ms365

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/juliankoehn/kspec/core"
)

type RiskyUserResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *RiskyUserResource) Name() string {
	return "ms365_risky_user"
}

func (r *RiskyUserResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	result, err := r.client.IdentityProtection().RiskyUsers().Get(ctx, nil)
	if err != nil {
		// Identity Protection might not be licensed
		return resources, nil
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, user := range result.GetValue() {
		data, err := json.Marshal(user)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = user.GetId()
		resourceMap["userDisplayName"] = user.GetUserDisplayName()
		resourceMap["userPrincipalName"] = user.GetUserPrincipalName()
		resourceMap["riskLevel"] = user.GetRiskLevel()
		resourceMap["riskState"] = user.GetRiskState()
		resourceMap["riskDetail"] = user.GetRiskDetail()
		resourceMap["riskLastUpdatedDateTime"] = user.GetRiskLastUpdatedDateTime()
		resourceMap["isDeleted"] = user.GetIsDeleted()
		resourceMap["isProcessing"] = user.GetIsProcessing()

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

type SecureScoreResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *SecureScoreResource) Name() string {
	return "ms365_secure_score"
}

func (r *SecureScoreResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	result, err := r.client.Security().SecureScores().Get(ctx, nil)
	if err != nil {
		// Security API might not be available
		return resources, nil
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	// Get the most recent secure score
	scores := result.GetValue()
	if len(scores) == 0 {
		return resources, nil
	}

	// Get the latest score (first one)
	score := scores[0]

	data, err := json.Marshal(score)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal secure score: %w", err)
	}

	var resourceMap map[string]interface{}
	if err := json.Unmarshal(data, &resourceMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secure score: %w", err)
	}

	resourceMap["id"] = score.GetId()
	resourceMap["currentScore"] = score.GetCurrentScore()
	resourceMap["maxScore"] = score.GetMaxScore()
	resourceMap["createdDateTime"] = score.GetCreatedDateTime()
	resourceMap["licensedUserCount"] = score.GetLicensedUserCount()
	resourceMap["activeUserCount"] = score.GetActiveUserCount()
	resourceMap["enabledServices"] = score.GetEnabledServices()

	// Calculate percentage
	if score.GetMaxScore() != nil && *score.GetMaxScore() > 0 {
		percentage := (*score.GetCurrentScore() / *score.GetMaxScore()) * 100
		resourceMap["scorePercentage"] = percentage
	}

	// Control scores
	if score.GetControlScores() != nil {
		var controls []map[string]interface{}
		for _, control := range score.GetControlScores() {
			controls = append(controls, map[string]interface{}{
				"controlName":        control.GetControlName(),
				"controlCategory":    control.GetControlCategory(),
				"score":              control.GetScore(),
				"description":        control.GetDescription(),
			})
		}
		resourceMap["controlScores"] = controls
	}

	resources = append(resources, resourceMap)
	return resources, nil
}

type SecurityAlertResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *SecurityAlertResource) Name() string {
	return "ms365_security_alert"
}

func (r *SecurityAlertResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	result, err := r.client.Security().Alerts_v2().Get(ctx, nil)
	if err != nil {
		// Security API might not be available
		return resources, nil
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, alert := range result.GetValue() {
		data, err := json.Marshal(alert)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = alert.GetId()
		resourceMap["title"] = alert.GetTitle()
		resourceMap["description"] = alert.GetDescription()
		resourceMap["severity"] = alert.GetSeverity()
		resourceMap["status"] = alert.GetStatus()
		resourceMap["createdDateTime"] = alert.GetCreatedDateTime()
		resourceMap["lastUpdateDateTime"] = alert.GetLastUpdateDateTime()
		resourceMap["classification"] = alert.GetClassification()
		resourceMap["determination"] = alert.GetDetermination()
		resourceMap["serviceSource"] = alert.GetServiceSource()
		resourceMap["detectorId"] = alert.GetDetectorId()

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
