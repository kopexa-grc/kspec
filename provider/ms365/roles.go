package ms365

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

// DirectoryRoleResource provides access to Azure AD directory roles.
type DirectoryRoleResource struct {
	client *msgraphsdk.GraphServiceClient
}

// Name returns the resource type identifier for directory roles.
func (r *DirectoryRoleResource) Name() string {
	return "ms365_directory_role"
}

// Fetch retrieves all directory roles with their members.
func (r *DirectoryRoleResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	result, err := r.client.DirectoryRoles().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory roles: %w", err)
	}

	roles := result.GetValue()
	if roles == nil {
		return []core.Resource{}, nil
	}

	resources := make([]core.Resource, 0, len(roles))
	for _, role := range roles {
		data, err := json.Marshal(role)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = role.GetId()
		resourceMap["displayName"] = role.GetDisplayName()
		resourceMap["description"] = role.GetDescription()
		resourceMap["roleTemplateId"] = role.GetRoleTemplateId()

		// Get members of this role
		members, err := r.client.DirectoryRoles().ByDirectoryRoleId(*role.GetId()).Members().Get(ctx, nil)
		if err == nil && members.GetValue() != nil {
			var memberList []map[string]interface{}
			for _, member := range members.GetValue() {
				memberList = append(memberList, map[string]interface{}{
					"id":   member.GetId(),
					"type": member.GetOdataType(),
				})
			}
			resourceMap["members"] = memberList
			resourceMap["memberCount"] = len(memberList)
		}

		// Check for privileged roles
		privilegedRoles := map[string]bool{
			"Global Administrator":             true,
			"Privileged Role Administrator":    true,
			"Security Administrator":           true,
			"User Administrator":               true,
			"Exchange Administrator":           true,
			"SharePoint Administrator":         true,
			"Teams Administrator":              true,
			"Application Administrator":        true,
			"Cloud Application Administrator":  true,
			"Authentication Administrator":     true,
			"Conditional Access Administrator": true,
		}
		resourceMap["isPrivilegedRole"] = privilegedRoles[*role.GetDisplayName()]

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

// RoleAssignmentResource provides access to Azure AD role assignments.
type RoleAssignmentResource struct {
	client *msgraphsdk.GraphServiceClient
}

// Name returns the resource type identifier for role assignments.
func (r *RoleAssignmentResource) Name() string {
	return "ms365_role_assignment"
}

// Fetch retrieves all role assignments from Azure AD role management.
func (r *RoleAssignmentResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// Get unified role assignments
	result, err := r.client.RoleManagement().Directory().RoleAssignments().Get(ctx, nil)
	if err != nil {
		// Role management might not be available
		return []core.Resource{}, err
	}

	assignments := result.GetValue()
	if assignments == nil {
		return []core.Resource{}, nil
	}

	resources := make([]core.Resource, 0, len(assignments))
	for _, assignment := range assignments {
		data, err := json.Marshal(assignment)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = assignment.GetId()
		resourceMap["roleDefinitionId"] = assignment.GetRoleDefinitionId()
		resourceMap["principalId"] = assignment.GetPrincipalId()
		resourceMap["directoryScopeId"] = assignment.GetDirectoryScopeId()

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
