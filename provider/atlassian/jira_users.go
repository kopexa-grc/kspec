package atlassian

import (
	"context"

	v3 "github.com/ctreminiom/go-atlassian/jira/v3"

	"github.com/kopexa-grc/kspec/core"
)

// Account type constants
const (
	AccountTypeAtlassian = "atlassian"
	AccountTypeApp       = "app"
	AccountTypeCustomer  = "customer"
)

// JiraUserResource fetches Jira users.
type JiraUserResource struct {
	client *v3.Client
}

// Name returns the resource type name.
func (r *JiraUserResource) Name() string {
	return "atlassian_jira_user"
}

// Fetch retrieves Jira users.
func (r *JiraUserResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	startAt := 0
	maxResults := 50

	for {
		// Search with empty query to get all users
		users, response, err := r.client.User.Search.Do(ctx, "", "", startAt, maxResults)
		if err != nil {
			if response != nil && response.Code == 404 {
				break
			}
			return nil, err
		}

		if len(users) == 0 {
			break
		}

		for _, user := range users {
			resource := make(core.Resource)
			resource["account_id"] = user.AccountID
			resource["account_type"] = user.AccountType
			resource["display_name"] = user.DisplayName
			resource["email_address"] = user.EmailAddress
			resource["active"] = user.Active
			resource["timezone"] = user.TimeZone
			resource["locale"] = user.Locale

			// Security flags
			resource["is_active"] = user.Active
			resource["is_atlassian_account"] = user.AccountType == AccountTypeAtlassian
			resource["is_app_account"] = user.AccountType == AccountTypeApp
			resource["is_customer_account"] = user.AccountType == AccountTypeCustomer
			resource["has_email"] = user.EmailAddress != ""

			resources = append(resources, resource)
		}

		if len(users) < maxResults {
			break
		}
		startAt += maxResults
	}

	return resources, nil
}
