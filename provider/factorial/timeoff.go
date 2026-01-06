package factorial

import (
	"context"
	"time"

	"github.com/kopexa-grc/kspec/core"
)

// LeaveResource fetches Factorial HR leave requests.
type LeaveResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *LeaveResource) Name() string {
	return "factorial_leave"
}

// Fetch retrieves all Factorial HR leave requests.
func (r *LeaveResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	leaves, err := r.conn.fetchPaginated(ctx, "resources/timeoff/leaves")
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(leaves))

	for _, leave := range leaves {
		resource := make(core.Resource)

		// Basic info
		if id, ok := leave["id"].(float64); ok {
			resource["id"] = int64(id)
		}
		if employeeID, ok := leave["employee_id"].(float64); ok {
			resource["employee_id"] = int64(employeeID)
		}
		if leaveTypeID, ok := leave["leave_type_id"].(float64); ok {
			resource["leave_type_id"] = int64(leaveTypeID)
		}

		resource["description"] = leave["description"]
		resource["start_on"] = leave["start_on"]
		resource["finish_on"] = leave["finish_on"]
		resource["half_day"] = leave["half_day"]
		resource["leave_type_name"] = leave["leave_type_name"]

		// Status
		resource["status"] = leave["status"]
		resource["approval_status"] = leave["approval_status"]

		// Approver info
		if approvedBy, ok := leave["approved_by"].(float64); ok {
			resource["approved_by"] = int64(approvedBy)
		}
		resource["approved_at"] = leave["approved_at"]

		// Timestamps
		resource["created_at"] = leave["created_at"]
		resource["updated_at"] = leave["updated_at"]

		// Computed fields for policy evaluation
		var status, approvalStatus string
		if v, ok := leave["status"].(string); ok {
			status = v
		}
		if v, ok := leave["approval_status"].(string); ok {
			approvalStatus = v
		}

		resource["is_approved"] = status == "approved" || approvalStatus == "approved"
		resource["is_rejected"] = status == "rejected" || approvalStatus == "rejected"
		resource["is_pending"] = status == "pending" || approvalStatus == "pending" || status == ""

		// Check if leave is in the past/future
		if startOn, ok := leave["start_on"].(string); ok {
			if t, err := time.Parse("2006-01-02", startOn); err == nil {
				resource["is_past"] = t.Before(time.Now().Truncate(24 * time.Hour))
				resource["is_future"] = t.After(time.Now())
			}
		}

		// Calculate duration
		if startOn, ok := leave["start_on"].(string); ok {
			if finishOn, ok := leave["finish_on"].(string); ok {
				if start, err := time.Parse("2006-01-02", startOn); err == nil {
					if finish, err := time.Parse("2006-01-02", finishOn); err == nil {
						days := int(finish.Sub(start).Hours()/24) + 1
						if halfDay, ok := leave["half_day"].(bool); ok && halfDay {
							resource["leave_days"] = 0.5
						} else {
							resource["leave_days"] = days
						}
					}
				}
			}
		}

		// Check if pending for too long (more than 7 days)
		if resource["is_pending"] == true {
			if createdAt, ok := leave["created_at"].(string); ok {
				if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
					daysPending := int(time.Since(t).Hours() / 24)
					resource["days_pending"] = daysPending
					resource["is_stale"] = daysPending > 7
				}
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// AllowanceResource fetches Factorial HR leave allowances.
type AllowanceResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *AllowanceResource) Name() string {
	return "factorial_allowance"
}

// Fetch retrieves all Factorial HR leave allowances.
func (r *AllowanceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	allowances, err := r.conn.fetchPaginated(ctx, "resources/timeoff/allowances")
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(allowances))

	for _, allowance := range allowances {
		resource := make(core.Resource)

		// Basic info
		if id, ok := allowance["id"].(float64); ok {
			resource["id"] = int64(id)
		}
		if employeeID, ok := allowance["employee_id"].(float64); ok {
			resource["employee_id"] = int64(employeeID)
		}
		if leaveTypeID, ok := allowance["leave_type_id"].(float64); ok {
			resource["leave_type_id"] = int64(leaveTypeID)
		}

		resource["leave_type_name"] = allowance["leave_type_name"]
		resource["allowance"] = allowance["allowance"]
		resource["taken"] = allowance["taken"]
		resource["remaining"] = allowance["remaining"]
		resource["balance"] = allowance["balance"]
		resource["pending"] = allowance["pending"]
		resource["extra"] = allowance["extra"]
		resource["carryover"] = allowance["carryover"]

		// Year info
		resource["year"] = allowance["year"]

		// Timestamps
		resource["created_at"] = allowance["created_at"]
		resource["updated_at"] = allowance["updated_at"]

		// Computed fields for policy evaluation
		if allowanceVal, ok := allowance["allowance"].(float64); ok {
			resource["has_allowance"] = allowanceVal > 0

			if remaining, ok := allowance["remaining"].(float64); ok {
				resource["utilization_rate"] = (allowanceVal - remaining) / allowanceVal * 100
				resource["is_exhausted"] = remaining <= 0
				resource["is_low"] = remaining > 0 && remaining < allowanceVal*0.1 // Less than 10% remaining
			}
		}

		if pending, ok := allowance["pending"].(float64); ok {
			resource["has_pending_requests"] = pending > 0
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
