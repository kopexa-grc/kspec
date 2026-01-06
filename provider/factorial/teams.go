package factorial

import (
	"context"

	"github.com/kopexa-grc/kspec/core"
)

// TeamResource fetches Factorial HR teams.
type TeamResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *TeamResource) Name() string {
	return "factorial_team"
}

// Fetch retrieves all Factorial HR teams.
func (r *TeamResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	teams, err := r.conn.fetchPaginated(ctx, "resources/teams/teams")
	if err != nil {
		return nil, err
	}

	// Fetch team memberships to count members and identify leads
	memberships, err := r.conn.fetchPaginated(ctx, "resources/teams/team_memberships")
	teamMemberCounts := make(map[float64]int)
	teamLeads := make(map[float64]bool)
	if err == nil {
		for _, m := range memberships {
			if teamID, ok := m["team_id"].(float64); ok {
				teamMemberCounts[teamID]++
				if isLead, ok := m["is_lead"].(bool); ok && isLead {
					teamLeads[teamID] = true
				}
			}
		}
	}

	resources := make([]core.Resource, 0, len(teams))

	for _, team := range teams {
		resource := make(core.Resource)

		// Basic info
		var teamID float64
		if id, ok := team["id"].(float64); ok {
			resource["id"] = int64(id)
			teamID = id
		}

		resource["name"] = team["name"]
		resource["description"] = team["description"]

		if companyID, ok := team["company_id"].(float64); ok {
			resource["company_id"] = int64(companyID)
		}

		// Employee IDs (if provided by API)
		if employeeIDs, ok := team["employee_ids"].([]any); ok {
			ids := make([]int64, 0, len(employeeIDs))
			for _, id := range employeeIDs {
				if fid, ok := id.(float64); ok {
					ids = append(ids, int64(fid))
				}
			}
			resource["employee_ids"] = ids
			resource["member_count"] = len(ids)
		} else {
			// Use membership count
			resource["member_count"] = teamMemberCounts[teamID]
		}

		if leadID, ok := team["lead_id"].(float64); ok {
			resource["lead_id"] = int64(leadID)
		}

		// Timestamps
		resource["created_at"] = team["created_at"]
		resource["updated_at"] = team["updated_at"]

		// Computed fields for policy evaluation
		var memberCount int
		if v, ok := resource["member_count"].(int); ok {
			memberCount = v
		}
		resource["has_members"] = teamMemberCounts[teamID] > 0 || memberCount > 0
		resource["has_lead"] = teamLeads[teamID] || team["lead_id"] != nil

		var teamName string
		if v, ok := team["name"].(string); ok {
			teamName = v
		}
		resource["has_name"] = teamName != ""

		var description string
		if v, ok := team["description"].(string); ok {
			description = v
		}
		resource["has_description"] = description != ""

		resources = append(resources, resource)
	}

	return resources, nil
}

// TeamMembershipResource fetches Factorial HR team memberships.
type TeamMembershipResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *TeamMembershipResource) Name() string {
	return "factorial_team_membership"
}

// Fetch retrieves all Factorial HR team memberships.
func (r *TeamMembershipResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	memberships, err := r.conn.fetchPaginated(ctx, "resources/teams/team_memberships")
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(memberships))

	for _, membership := range memberships {
		resource := make(core.Resource)

		// Basic info
		if id, ok := membership["id"].(float64); ok {
			resource["id"] = int64(id)
		}
		if teamID, ok := membership["team_id"].(float64); ok {
			resource["team_id"] = int64(teamID)
		}
		if employeeID, ok := membership["employee_id"].(float64); ok {
			resource["employee_id"] = int64(employeeID)
		}

		// Role/Lead status
		resource["is_lead"] = membership["is_lead"]
		resource["role"] = membership["role"]

		// Timestamps
		resource["created_at"] = membership["created_at"]
		resource["updated_at"] = membership["updated_at"]

		resources = append(resources, resource)
	}

	return resources, nil
}
