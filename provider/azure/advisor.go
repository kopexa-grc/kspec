// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/advisor/armadvisor"

	"github.com/kopexa-grc/kspec/core"
)

// AdvisorRecommendationResource represents an Azure Advisor recommendation resource scanner.
type AdvisorRecommendationResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *AdvisorRecommendationResource) Name() string {
	return "azure_advisor_recommendation"
}

// Fetch retrieves Advisor recommendations from Azure.
func (r *AdvisorRecommendationResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armadvisor.NewRecommendationsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Advisor client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Advisor recommendations: %w", err)
		}

		for _, rec := range page.Value {
			data, err := json.Marshal(rec)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			if rec.Properties != nil {
				resourceMap["category"] = rec.Properties.Category
				resourceMap["impact"] = rec.Properties.Impact
				resourceMap["impactedField"] = rec.Properties.ImpactedField
				resourceMap["impactedValue"] = rec.Properties.ImpactedValue
				resourceMap["lastUpdated"] = rec.Properties.LastUpdated
				resourceMap["risk"] = rec.Properties.Risk

				if rec.Properties.ShortDescription != nil {
					resourceMap["problem"] = rec.Properties.ShortDescription.Problem
					resourceMap["solution"] = rec.Properties.ShortDescription.Solution
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
