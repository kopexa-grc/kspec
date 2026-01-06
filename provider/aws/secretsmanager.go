package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/kopexa-grc/kspec/core"
)

// SecretsManagerSecretResource fetches Secrets Manager secrets.
type SecretsManagerSecretResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *SecretsManagerSecretResource) Name() string {
	return "aws_secretsmanager_secret"
}

// Fetch retrieves all Secrets Manager secrets across configured regions.
func (r *SecretsManagerSecretResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := secretsmanager.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := secretsmanager.NewListSecretsPaginator(client, &secretsmanager.ListSecretsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, secret := range page.SecretList {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(secret.Name)
				resource["arn"] = aws.ToString(secret.ARN)
				resource["name"] = aws.ToString(secret.Name)
				resource["region"] = region

				resource["description"] = aws.ToString(secret.Description)

				// Encryption
				resource["kms_key_id"] = aws.ToString(secret.KmsKeyId)
				resource["has_cmk_encryption"] = secret.KmsKeyId != nil && aws.ToString(secret.KmsKeyId) != ""

				// Rotation
				resource["rotation_enabled"] = secret.RotationEnabled != nil && *secret.RotationEnabled
				resource["has_rotation"] = secret.RotationEnabled != nil && *secret.RotationEnabled
				if secret.RotationLambdaARN != nil {
					resource["rotation_lambda_arn"] = aws.ToString(secret.RotationLambdaARN)
				}
				if secret.RotationRules != nil {
					if secret.RotationRules.AutomaticallyAfterDays != nil {
						resource["rotation_days"] = *secret.RotationRules.AutomaticallyAfterDays
					}
					resource["rotation_schedule"] = aws.ToString(secret.RotationRules.ScheduleExpression)
				}

				// Last rotation
				if secret.LastRotatedDate != nil {
					resource["last_rotated_date"] = secret.LastRotatedDate.String()
					daysSinceRotation := int(time.Since(*secret.LastRotatedDate).Hours() / 24)
					resource["days_since_rotation"] = daysSinceRotation
					resource["needs_rotation"] = daysSinceRotation > 90 // Common best practice
				} else {
					resource["needs_rotation"] = true // Never rotated
				}

				// Last accessed
				if secret.LastAccessedDate != nil {
					resource["last_accessed_date"] = secret.LastAccessedDate.String()
					daysSinceAccess := int(time.Since(*secret.LastAccessedDate).Hours() / 24)
					resource["days_since_access"] = daysSinceAccess
					resource["is_unused"] = daysSinceAccess > 90
				}

				// Last changed
				if secret.LastChangedDate != nil {
					resource["last_changed_date"] = secret.LastChangedDate.String()
				}

				// Creation date
				if secret.CreatedDate != nil {
					resource["created_date"] = secret.CreatedDate.String()
				}

				// Deletion
				if secret.DeletedDate != nil {
					resource["deleted_date"] = secret.DeletedDate.String()
					resource["is_deleted"] = true
				} else {
					resource["is_deleted"] = false
				}

				// Primary region (for replication)
				resource["primary_region"] = aws.ToString(secret.PrimaryRegion)
				resource["is_replicated"] = secret.PrimaryRegion != nil && aws.ToString(secret.PrimaryRegion) != ""

				// Owning service
				resource["owning_service"] = aws.ToString(secret.OwningService)
				resource["is_aws_managed"] = secret.OwningService != nil && aws.ToString(secret.OwningService) != ""

				// Secret versions
				if secret.SecretVersionsToStages != nil {
					resource["version_count"] = len(secret.SecretVersionsToStages)
					resource["has_multiple_versions"] = len(secret.SecretVersionsToStages) > 1
				}

				// Tags
				tags := make(map[string]string)
				for _, tag := range secret.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				// Get resource policy
				policyResp, err := client.GetResourcePolicy(ctx, &secretsmanager.GetResourcePolicyInput{
					SecretId: secret.ARN,
				})
				if err == nil && policyResp.ResourcePolicy != nil {
					resource["has_resource_policy"] = true
					resource["resource_policy"] = aws.ToString(policyResp.ResourcePolicy)
				} else {
					resource["has_resource_policy"] = false
				}

				// Computed
				rotationEnabled, _ := resource["rotation_enabled"].(bool) //nolint:errcheck // zero value ok
				hasCMK, _ := resource["has_cmk_encryption"].(bool)        //nolint:errcheck // zero value ok
				resource["meets_security_baseline"] = rotationEnabled && hasCMK

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
