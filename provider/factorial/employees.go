package factorial

import (
	"context"
	"time"

	"github.com/kopexa-grc/kspec/core"
)

// EmployeeResource fetches Factorial HR employees.
type EmployeeResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *EmployeeResource) Name() string {
	return "factorial_employee"
}

// Fetch retrieves all Factorial HR employees.
func (r *EmployeeResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	employees, err := r.conn.fetchPaginated(ctx, "resources/employees/employees")
	if err != nil {
		return nil, err
	}

	// Fetch teams to check team membership
	teamMembers, err := r.conn.fetchPaginated(ctx, "resources/teams/team_memberships")
	employeeTeams := make(map[float64]bool)
	if err == nil {
		for _, tm := range teamMembers {
			if empID, ok := tm["employee_id"].(float64); ok {
				employeeTeams[empID] = true
			}
		}
	}

	resources := make([]core.Resource, 0, len(employees))

	for _, emp := range employees {
		resource := make(core.Resource)

		// Basic info
		if id, ok := emp["id"].(float64); ok {
			resource["id"] = int64(id)
		}
		resource["first_name"] = emp["first_name"]
		resource["last_name"] = emp["last_name"]

		var firstName, lastName string
		if v, ok := emp["first_name"].(string); ok {
			firstName = v
		}
		if v, ok := emp["last_name"].(string); ok {
			lastName = v
		}
		resource["full_name"] = firstName + " " + lastName

		resource["email"] = emp["email"]
		resource["identifier"] = emp["identifier"]
		resource["identifier_type"] = emp["identifier_type"]
		resource["gender"] = emp["gender"]
		resource["nationality"] = emp["nationality"]
		resource["bank_number"] = emp["bank_number"]
		resource["country"] = emp["country"]
		resource["city"] = emp["city"]
		resource["state"] = emp["state"]
		resource["postal_code"] = emp["postal_code"]
		resource["address_line_1"] = emp["address_line_1"]
		resource["address_line_2"] = emp["address_line_2"]
		resource["swift_bic"] = emp["swift_bic"]
		resource["phone_number"] = emp["phone_number"]
		resource["social_security_number"] = emp["social_security_number"]

		// Dates
		resource["birthday_on"] = emp["birthday_on"]
		resource["start_date"] = emp["start_date"]
		resource["termination_date"] = emp["termination_date"]
		resource["terminated_on"] = emp["terminated_on"]

		// IDs and relationships
		if managerID, ok := emp["manager_id"].(float64); ok {
			resource["manager_id"] = int64(managerID)
		}
		if locationID, ok := emp["location_id"].(float64); ok {
			resource["location_id"] = int64(locationID)
		}
		if companyID, ok := emp["company_id"].(float64); ok {
			resource["company_id"] = int64(companyID)
		}
		if legalEntityID, ok := emp["legal_entity_id"].(float64); ok {
			resource["legal_entity_id"] = int64(legalEntityID)
		}
		if timeoffManagerID, ok := emp["timeoff_manager_id"].(float64); ok {
			resource["timeoff_manager_id"] = int64(timeoffManagerID)
		}

		// Status flags from API
		resource["invitation_pending"] = emp["invitation_pending"]
		resource["is_manager"] = emp["is_manager"]

		// Timestamps
		resource["created_at"] = emp["created_at"]
		resource["updated_at"] = emp["updated_at"]

		// Computed security/compliance fields
		var email string
		if v, ok := emp["email"].(string); ok {
			email = v
		}
		resource["has_email"] = email != ""

		managerID, hasManager := emp["manager_id"].(float64)
		resource["has_manager"] = hasManager && managerID > 0

		// Check if employee is active (no termination date or termination date in future)
		isActive := true
		if terminatedOn, ok := emp["terminated_on"].(string); ok && terminatedOn != "" {
			if t, err := time.Parse("2006-01-02", terminatedOn); err == nil {
				isActive = t.After(time.Now())
			} else {
				isActive = false
			}
		}
		resource["is_active"] = isActive

		// Check team membership
		var empID float64
		if v, ok := emp["id"].(float64); ok {
			empID = v
		}
		resource["has_team"] = employeeTeams[empID]

		// Calculate employment days
		if startDate, ok := emp["start_date"].(string); ok && startDate != "" {
			if t, err := time.Parse("2006-01-02", startDate); err == nil {
				resource["employment_days"] = int(time.Since(t).Hours() / 24)
			}
		}

		// Check if invitation is pending
		var invitationPending bool
		if v, ok := emp["invitation_pending"].(bool); ok {
			invitationPending = v
		}
		resource["has_pending_invitation"] = invitationPending

		// Has labels/custom fields
		if customFields, ok := emp["custom_fields"].(map[string]any); ok {
			resource["custom_fields"] = customFields
			resource["has_custom_fields"] = len(customFields) > 0
		} else {
			resource["has_custom_fields"] = false
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
