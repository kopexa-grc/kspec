package factorial

import (
	"context"

	"github.com/kopexa-grc/kspec/core"
)

// LocationResource fetches Factorial HR locations.
type LocationResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *LocationResource) Name() string {
	return "factorial_location"
}

// Fetch retrieves all Factorial HR locations.
func (r *LocationResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	locations, err := r.conn.fetchPaginated(ctx, "resources/locations/locations")
	if err != nil {
		return nil, err
	}

	// Count employees per location
	employees, err := r.conn.fetchPaginated(ctx, "resources/employees/employees")
	locationEmployeeCounts := make(map[float64]int)
	if err == nil {
		for _, emp := range employees {
			if locationID, ok := emp["location_id"].(float64); ok {
				locationEmployeeCounts[locationID]++
			}
		}
	}

	resources := make([]core.Resource, 0, len(locations))

	for _, location := range locations {
		resource := make(core.Resource)

		// Basic info
		var locationID float64
		if id, ok := location["id"].(float64); ok {
			resource["id"] = int64(id)
			locationID = id
		}
		if companyID, ok := location["company_id"].(float64); ok {
			resource["company_id"] = int64(companyID)
		}

		resource["name"] = location["name"]
		resource["country"] = location["country"]
		resource["state"] = location["state"]
		resource["city"] = location["city"]
		resource["address_line_1"] = location["address_line_1"]
		resource["address_line_2"] = location["address_line_2"]
		resource["postal_code"] = location["postal_code"]
		resource["phone_number"] = location["phone_number"]

		// Timezone
		resource["timezone"] = location["timezone"]

		// Main location flag
		resource["is_main"] = location["is_main"]

		// Timestamps
		resource["created_at"] = location["created_at"]
		resource["updated_at"] = location["updated_at"]

		// Computed fields for policy evaluation
		var locName string
		if v, ok := location["name"].(string); ok {
			locName = v
		}
		resource["has_name"] = locName != ""

		var country string
		if v, ok := location["country"].(string); ok {
			country = v
		}
		resource["has_country"] = country != ""

		var address string
		if v, ok := location["address_line_1"].(string); ok {
			address = v
		}
		resource["has_address"] = address != ""

		var timezone string
		if v, ok := location["timezone"].(string); ok {
			timezone = v
		}
		resource["has_timezone"] = timezone != ""

		// Employee count
		resource["employee_count"] = locationEmployeeCounts[locationID]
		resource["has_employees"] = locationEmployeeCounts[locationID] > 0

		// Check if main office
		var isMain bool
		if v, ok := location["is_main"].(bool); ok {
			isMain = v
		}
		resource["is_main_location"] = isMain

		resources = append(resources, resource)
	}

	return resources, nil
}
