package factorial

import (
	"context"
	"time"

	"github.com/kopexa-grc/kspec/core"
)

// ShiftResource fetches Factorial HR attendance shifts.
type ShiftResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *ShiftResource) Name() string {
	return "factorial_shift"
}

// Fetch retrieves all Factorial HR shifts.
func (r *ShiftResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	shifts, err := r.conn.fetchPaginated(ctx, "resources/attendance/shifts")
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(shifts))

	for _, shift := range shifts {
		resource := make(core.Resource)

		// Basic info
		if id, ok := shift["id"].(float64); ok {
			resource["id"] = int64(id)
		}
		if employeeID, ok := shift["employee_id"].(float64); ok {
			resource["employee_id"] = int64(employeeID)
		}

		resource["day"] = shift["day"]
		resource["date"] = shift["date"]
		resource["clock_in"] = shift["clock_in"]
		resource["clock_out"] = shift["clock_out"]
		resource["minutes"] = shift["minutes"]
		resource["observations"] = shift["observations"]
		resource["half_day"] = shift["half_day"]

		// Location info
		if locationID, ok := shift["location_id"].(float64); ok {
			resource["location_id"] = int64(locationID)
		}
		resource["location_type"] = shift["location_type"]

		// Timestamps
		resource["created_at"] = shift["created_at"]
		resource["updated_at"] = shift["updated_at"]

		// Computed fields for policy evaluation
		clockIn, hasClockIn := shift["clock_in"].(string)
		clockOut, hasClockOut := shift["clock_out"].(string)

		resource["has_clock_in"] = hasClockIn && clockIn != ""
		resource["has_clock_out"] = hasClockOut && clockOut != ""
		resource["is_complete"] = (hasClockIn && clockIn != "") && (hasClockOut && clockOut != "")

		// Calculate duration if both clock times exist
		if hasClockIn && hasClockOut && clockIn != "" && clockOut != "" {
			// Try to parse times and calculate duration
			if inTime, err := time.Parse("15:04", clockIn); err == nil {
				if outTime, err := time.Parse("15:04", clockOut); err == nil {
					duration := outTime.Sub(inTime)
					resource["duration_hours"] = duration.Hours()
					resource["is_overtime"] = duration.Hours() > 8
					resource["is_short_shift"] = duration.Hours() < 4
				}
			}
		}

		// Check if shift is today
		if date, ok := shift["date"].(string); ok {
			if t, err := time.Parse("2006-01-02", date); err == nil {
				resource["is_today"] = t.Format("2006-01-02") == time.Now().Format("2006-01-02")
				resource["is_past"] = t.Before(time.Now().Truncate(24 * time.Hour))
			}
		}

		// Minutes worked
		if minutes, ok := shift["minutes"].(float64); ok {
			resource["worked_minutes"] = int(minutes)
			resource["worked_hours"] = minutes / 60
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
