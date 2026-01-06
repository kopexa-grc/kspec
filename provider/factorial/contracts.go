package factorial

import (
	"context"
	"time"

	"github.com/kopexa-grc/kspec/core"
)

// ContractResource fetches Factorial HR contract versions.
type ContractResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *ContractResource) Name() string {
	return "factorial_contract"
}

// Fetch retrieves all Factorial HR contract versions.
func (r *ContractResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	contracts, err := r.conn.fetchPaginated(ctx, "resources/contracts/contract_versions")
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(contracts))

	for _, contract := range contracts {
		resource := make(core.Resource)

		// Basic info
		if id, ok := contract["id"].(float64); ok {
			resource["id"] = int64(id)
		}
		if employeeID, ok := contract["employee_id"].(float64); ok {
			resource["employee_id"] = int64(employeeID)
		}

		resource["job_title"] = contract["job_title"]
		resource["role"] = contract["role"]
		resource["level"] = contract["level"]
		resource["working_week_days"] = contract["working_week_days"]
		resource["working_hours"] = contract["working_hours"]
		resource["working_hours_frequency"] = contract["working_hours_frequency"]
		resource["salary_amount"] = contract["salary_amount"]
		resource["salary_frequency"] = contract["salary_frequency"]
		resource["salary_currency"] = contract["salary_currency"]
		resource["country"] = contract["country"]
		resource["has_payroll"] = contract["has_payroll"]
		resource["es_has_teleworking_contract"] = contract["es_has_teleworking_contract"]
		resource["es_cotization_group"] = contract["es_cotization_group"]
		resource["es_contract_type_id"] = contract["es_contract_type_id"]
		resource["es_job_description"] = contract["es_job_description"]
		resource["es_working_day_type_id"] = contract["es_working_day_type_id"]
		resource["fr_employee_type"] = contract["fr_employee_type"]
		resource["fr_forfait_jours"] = contract["fr_forfait_jours"]
		resource["fr_professional_category_id"] = contract["fr_professional_category_id"]
		resource["fr_work_type"] = contract["fr_work_type"]
		resource["de_contract_type_id"] = contract["de_contract_type_id"]
		resource["pt_contract_type_id"] = contract["pt_contract_type_id"]
		resource["pt_work_type"] = contract["pt_work_type"]
		resource["job_catalog_role_id"] = contract["job_catalog_role_id"]

		// Dates
		resource["effective_on"] = contract["effective_on"]
		resource["starts_on"] = contract["starts_on"]
		resource["ends_on"] = contract["ends_on"]

		// Timestamps
		resource["created_at"] = contract["created_at"]
		resource["updated_at"] = contract["updated_at"]

		// Computed fields for policy evaluation
		jobTitle, _ := contract["job_title"].(string)
		resource["has_job_title"] = jobTitle != ""

		// Check if contract is active (no end date or end date in future)
		isActive := true
		if endsOn, ok := contract["ends_on"].(string); ok && endsOn != "" {
			if t, err := time.Parse("2006-01-02", endsOn); err == nil {
				isActive = t.After(time.Now())
			} else {
				isActive = false
			}
		}
		resource["is_active"] = isActive

		// Check if contract is current (effective date has passed)
		isCurrent := false
		if effectiveOn, ok := contract["effective_on"].(string); ok && effectiveOn != "" {
			if t, err := time.Parse("2006-01-02", effectiveOn); err == nil {
				isCurrent = !t.After(time.Now())
			}
		}
		resource["is_current"] = isCurrent

		// Check if it's a permanent contract (no end date)
		_, hasEndDate := contract["ends_on"].(string)
		resource["is_permanent"] = !hasEndDate || contract["ends_on"] == "" || contract["ends_on"] == nil

		// Working hours validation
		if workingHours, ok := contract["working_hours"].(float64); ok {
			resource["has_working_hours"] = workingHours > 0
			resource["is_full_time"] = workingHours >= 35
		} else {
			resource["has_working_hours"] = false
			resource["is_full_time"] = false
		}

		// Salary info validation
		if salaryAmount, ok := contract["salary_amount"].(float64); ok {
			resource["has_salary"] = salaryAmount > 0
		} else {
			resource["has_salary"] = false
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
