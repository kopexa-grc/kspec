// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

// Team provides access to Microsoft Teams resources.
type Team struct {
	client *msgraphsdk.GraphServiceClient
}

// NewTeam creates a new Team resource.
func NewTeam(client *msgraphsdk.GraphServiceClient) *Team {
	return &Team{client: client}
}

// Name returns the resource type identifier for Microsoft Teams.
func (r *Team) Name() string {
	return "ms365_team"
}

// Fetch retrieves all Microsoft Teams with their settings and membership details.
func (r *Team) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// Get all teams (groups with Teams enabled)
	result, err := r.client.Teams().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}

	if result.GetValue() == nil {
		return nil, nil
	}

	resources := make([]core.Resource, 0, len(result.GetValue()))
	for _, team := range result.GetValue() {
		data, err := json.Marshal(team)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = team.GetId()
		resourceMap["displayName"] = team.GetDisplayName()
		resourceMap["description"] = team.GetDescription()
		resourceMap["visibility"] = team.GetVisibility()
		resourceMap["isArchived"] = team.GetIsArchived()
		resourceMap["webUrl"] = team.GetWebUrl()
		resourceMap["createdDateTime"] = team.GetCreatedDateTime()

		// Guest settings
		if team.GetGuestSettings() != nil {
			guest := team.GetGuestSettings()
			resourceMap["guestSettings"] = map[string]interface{}{
				"allowCreateUpdateChannels": guest.GetAllowCreateUpdateChannels(),
				"allowDeleteChannels":       guest.GetAllowDeleteChannels(),
			}
		}

		// Member settings
		if team.GetMemberSettings() != nil {
			member := team.GetMemberSettings()
			resourceMap["memberSettings"] = map[string]interface{}{
				"allowCreateUpdateChannels":         member.GetAllowCreateUpdateChannels(),
				"allowDeleteChannels":               member.GetAllowDeleteChannels(),
				"allowAddRemoveApps":                member.GetAllowAddRemoveApps(),
				"allowCreateUpdateRemoveTabs":       member.GetAllowCreateUpdateRemoveTabs(),
				"allowCreateUpdateRemoveConnectors": member.GetAllowCreateUpdateRemoveConnectors(),
			}
		}

		// Messaging settings
		if team.GetMessagingSettings() != nil {
			messaging := team.GetMessagingSettings()
			resourceMap["messagingSettings"] = map[string]interface{}{
				"allowUserEditMessages":    messaging.GetAllowUserEditMessages(),
				"allowUserDeleteMessages":  messaging.GetAllowUserDeleteMessages(),
				"allowOwnerDeleteMessages": messaging.GetAllowOwnerDeleteMessages(),
				"allowTeamMentions":        messaging.GetAllowTeamMentions(),
				"allowChannelMentions":     messaging.GetAllowChannelMentions(),
			}
		}

		// Fun settings
		if team.GetFunSettings() != nil {
			fun := team.GetFunSettings()
			resourceMap["funSettings"] = map[string]interface{}{
				"allowGiphy":            fun.GetAllowGiphy(),
				"giphyContentRating":    fun.GetGiphyContentRating(),
				"allowStickersAndMemes": fun.GetAllowStickersAndMemes(),
				"allowCustomMemes":      fun.GetAllowCustomMemes(),
			}
		}

		// Get channels
		channels, err := r.client.Teams().ByTeamId(*team.GetId()).Channels().Get(ctx, nil)
		if err == nil && channels.GetValue() != nil {
			var channelList []map[string]interface{}
			for _, channel := range channels.GetValue() {
				channelList = append(channelList, map[string]interface{}{
					"id":             channel.GetId(),
					"displayName":    channel.GetDisplayName(),
					"description":    channel.GetDescription(),
					"membershipType": channel.GetMembershipType(),
				})
			}
			resourceMap["channels"] = channelList
			resourceMap["channelCount"] = len(channelList)
		}

		// Get members count
		members, err := r.client.Teams().ByTeamId(*team.GetId()).Members().Get(ctx, nil)
		if err == nil && members.GetValue() != nil {
			resourceMap["memberCount"] = len(members.GetValue())

			// Count owners
			ownerCount := 0
			for _, member := range members.GetValue() {
				roles := member.GetRoles()
				for _, role := range roles {
					if role == "owner" {
						ownerCount++
						break
					}
				}
			}
			resourceMap["ownerCount"] = ownerCount
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

// TeamSettings provides access to tenant-wide Microsoft Teams settings.
type TeamSettings struct {
	client *msgraphsdk.GraphServiceClient
}

// NewTeamSettings creates a new TeamSettings resource.
func NewTeamSettings(client *msgraphsdk.GraphServiceClient) *TeamSettings {
	return &TeamSettings{client: client}
}

// Name returns the resource type identifier for Teams settings.
func (r *TeamSettings) Name() string {
	return "ms365_teams_settings"
}

// Fetch retrieves tenant-wide Microsoft Teams configuration settings.
func (r *TeamSettings) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	resources := make([]core.Resource, 0, 1)

	// Get tenant-wide Teams settings (requires Teams admin role)
	// This is obtained through teamwork settings
	result, err := r.client.Teamwork().Get(ctx, nil)
	if err != nil {
		return resources, err
	}

	resourceMap := make(map[string]interface{})
	resourceMap["id"] = result.GetId()

	// Check if device usage is available
	if result.GetWorkforceIntegrations() != nil {
		resourceMap["workforceIntegrationsCount"] = len(result.GetWorkforceIntegrations())
	}

	resources = append(resources, resourceMap)
	return resources, nil
}
