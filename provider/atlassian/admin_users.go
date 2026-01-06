package atlassian

import (
	"context"

	"github.com/ctreminiom/go-atlassian/admin"
	"github.com/kopexa-grc/kspec/core"
)

// AdminUserResource fetches organization users via Admin API.
type AdminUserResource struct {
	client *admin.Client
	orgID  string
}

// Name returns the resource type name.
func (r *AdminUserResource) Name() string {
	return "atlassian_admin_user"
}

// Fetch retrieves organization users.
func (r *AdminUserResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.orgID == "" {
		return nil, nil
	}

	var resources []core.Resource
	cursor := ""

	for {
		users, response, err := r.client.Organization.Users(ctx, r.orgID, cursor)
		if err != nil {
			if response != nil && response.Code == 404 {
				break
			}
			return nil, err
		}

		for _, user := range users.Data {
			resource := make(core.Resource)
			resource["account_id"] = user.AccountID
			resource["account_type"] = user.AccountType
			resource["account_status"] = user.AccountStatus
			resource["name"] = user.Name
			resource["email"] = user.Email
			resource["access_billable"] = user.AccessBillable
			resource["last_active"] = user.LastActive

			// Product access
			var products []string
			for _, product := range user.ProductAccess {
				if product != nil && product.Name != "" {
					products = append(products, product.Name)
				}
			}
			resource["products"] = products
			resource["product_count"] = len(user.ProductAccess)

			// Security flags
			resource["is_active"] = user.AccountStatus == "active"
			resource["is_inactive"] = user.AccountStatus == "inactive"
			resource["is_closed"] = user.AccountStatus == "closed"
			resource["is_billable"] = user.AccessBillable
			resource["has_email"] = user.Email != ""
			resource["has_product_access"] = len(user.ProductAccess) > 0

			resources = append(resources, resource)
		}

		// Check for next page
		if users.Links == nil || users.Links.Next == "" {
			break
		}
		cursor = users.Links.Next

		// Limit to 1000 users to avoid rate limiting
		if len(resources) >= 1000 {
			break
		}
	}

	return resources, nil
}
