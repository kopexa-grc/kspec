package ms365

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kopexa-grc/kspec/core"
)

// DeviceManagementItem represents the common interface for MS Graph device management items.
type DeviceManagementItem interface {
	GetId() *string
	GetDisplayName() *string
	GetDescription() *string
	GetCreatedDateTime() *time.Time
	GetLastModifiedDateTime() *time.Time
	GetVersion() *int32
	GetOdataType() *string
	GetAdditionalData() map[string]any
}

// fetchDeviceManagementResources is a generic helper for fetching MS365 device management resources.
// It handles the common pattern of iterating through results and converting them to core.Resource.
func fetchDeviceManagementResources[T DeviceManagementItem](
	ctx context.Context,
	items []T,
	typeFieldName string,
) []core.Resource {
	resources := make([]core.Resource, 0, len(items))

	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = item.GetId()
		resourceMap["displayName"] = item.GetDisplayName()
		resourceMap["description"] = item.GetDescription()
		resourceMap["createdDateTime"] = item.GetCreatedDateTime()
		resourceMap["lastModifiedDateTime"] = item.GetLastModifiedDateTime()
		resourceMap["version"] = item.GetVersion()

		// The @odata.type tells us what kind of configuration/policy this is
		if odataType := item.GetOdataType(); odataType != nil {
			resourceMap[typeFieldName] = *odataType
		}

		// Get additional data which contains platform-specific settings
		for key, value := range item.GetAdditionalData() {
			resourceMap[key] = value
		}

		resources = append(resources, resourceMap)
	}

	return resources
}
