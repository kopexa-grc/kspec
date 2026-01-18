// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ptr"
)

// KMSClient is the interface for KMS operations.
//
//nolint:dupl // AWS SDK client interfaces have similar structures by design
type KMSClient interface {
	ListKeys(ctx context.Context, params *kms.ListKeysInput, optFns ...func(*kms.Options)) (*kms.ListKeysOutput, error)
	DescribeKey(ctx context.Context, params *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error)
	GetKeyRotationStatus(ctx context.Context, params *kms.GetKeyRotationStatusInput, optFns ...func(*kms.Options)) (*kms.GetKeyRotationStatusOutput, error)
	GetKeyPolicy(ctx context.Context, params *kms.GetKeyPolicyInput, optFns ...func(*kms.Options)) (*kms.GetKeyPolicyOutput, error)
	ListAliases(ctx context.Context, params *kms.ListAliasesInput, optFns ...func(*kms.Options)) (*kms.ListAliasesOutput, error)
	ListGrants(ctx context.Context, params *kms.ListGrantsInput, optFns ...func(*kms.Options)) (*kms.ListGrantsOutput, error)
	ListResourceTags(ctx context.Context, params *kms.ListResourceTagsInput, optFns ...func(*kms.Options)) (*kms.ListResourceTagsOutput, error)
}

// KMSClientFactory creates KMS clients for specific regions.
type KMSClientFactory func(region string) KMSClient

// KMSKey fetches KMS keys.
type KMSKey struct {
	clientFactory KMSClientFactory
	accountID     string
	regions       []string
}

// NewKMSKey creates a new KMSKey resource.
func NewKMSKey(clientFactory KMSClientFactory, accountID string, regions []string) *KMSKey {
	return &KMSKey{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *KMSKey) Name() string {
	return "aws_kms_key"
}

// Fetch retrieves all KMS keys across configured regions.
func (r *KMSKey) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := kms.NewListKeysPaginator(client, &kms.ListKeysInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, key := range page.Keys {
				keyID := aws.ToString(key.KeyId)

				// Get key details
				keyResp, err := client.DescribeKey(ctx, &kms.DescribeKeyInput{
					KeyId: key.KeyId,
				})
				if err != nil {
					continue
				}

				metadata := keyResp.KeyMetadata
				if metadata == nil {
					continue
				}

				// AWS managed keys are included but marked accordingly

				resource := make(core.Resource)
				resource["id"] = keyID
				resource["region"] = region
				resource["arn"] = aws.ToString(metadata.Arn)

				resource["key_id"] = keyID
				resource["description"] = aws.ToString(metadata.Description)
				resource["key_state"] = string(metadata.KeyState)
				resource["key_usage"] = string(metadata.KeyUsage)
				resource["key_spec"] = string(metadata.KeySpec)
				resource["origin"] = string(metadata.Origin)
				resource["key_manager"] = string(metadata.KeyManager)

				// Is enabled
				enabled := metadata.Enabled
				resource["enabled"] = enabled
				resource["is_enabled"] = enabled

				// AWS managed vs customer managed
				resource["is_aws_managed"] = metadata.KeyManager == types.KeyManagerTypeAws
				resource["is_customer_managed"] = metadata.KeyManager == types.KeyManagerTypeCustomer

				// Multi-region
				resource["multi_region"] = ptr.Deref(metadata.MultiRegion, false)

				// Creation date
				if metadata.CreationDate != nil {
					resource["created_at"] = metadata.CreationDate.String()
				}

				// Deletion date (if scheduled)
				if metadata.DeletionDate != nil {
					resource["deletion_date"] = metadata.DeletionDate.String()
					resource["pending_deletion"] = true
				} else {
					resource["pending_deletion"] = false
				}

				// Valid until (for imported keys)
				if metadata.ValidTo != nil {
					resource["valid_to"] = metadata.ValidTo.String()
				}

				// Check key rotation status (only for symmetric keys)
				if metadata.KeyUsage == types.KeyUsageTypeEncryptDecrypt && metadata.KeySpec == types.KeySpecSymmetricDefault {
					rotationResp, err := client.GetKeyRotationStatus(ctx, &kms.GetKeyRotationStatusInput{
						KeyId: key.KeyId,
					})
					if err == nil {
						resource["key_rotation_enabled"] = rotationResp.KeyRotationEnabled
					} else {
						resource["key_rotation_enabled"] = false
					}
				} else {
					resource["key_rotation_enabled"] = false
				}

				// Get key policy
				policyResp, err := client.GetKeyPolicy(ctx, &kms.GetKeyPolicyInput{
					KeyId:      key.KeyId,
					PolicyName: aws.String("default"),
				})
				if err == nil && policyResp.Policy != nil {
					policyDoc := aws.ToString(policyResp.Policy)
					resource["key_policy"] = policyDoc

					// Analyze policy
					var policy map[string]any
					if err := json.Unmarshal([]byte(policyDoc), &policy); err == nil {
						hasPublicAccess, allowsCrossAccount := analyzeKMSPolicy(policy, r.accountID)
						resource["has_public_access"] = hasPublicAccess
						resource["allows_cross_account"] = allowsCrossAccount
					}
				}

				// Get aliases
				aliasResp, err := client.ListAliases(ctx, &kms.ListAliasesInput{
					KeyId: key.KeyId,
				})
				if err == nil {
					aliases := make([]string, 0, len(aliasResp.Aliases))
					for _, alias := range aliasResp.Aliases {
						aliases = append(aliases, aws.ToString(alias.AliasName))
					}
					resource["aliases"] = aliases
					resource["alias_count"] = len(aliases)

					// Get primary alias (first non-aws alias)
					for _, alias := range aliases {
						if !strings.HasPrefix(alias, "alias/aws/") {
							resource["primary_alias"] = alias
							break
						}
					}
				}

				// Get grants
				grantsResp, err := client.ListGrants(ctx, &kms.ListGrantsInput{
					KeyId: key.KeyId,
				})
				if err == nil {
					resource["grant_count"] = len(grantsResp.Grants)
					resource["has_grants"] = len(grantsResp.Grants) > 0
				}

				// Get tags
				tagsResp, err := client.ListResourceTags(ctx, &kms.ListResourceTagsInput{
					KeyId: key.KeyId,
				})
				if err == nil {
					tags := make(map[string]string)
					for _, tag := range tagsResp.Tags {
						tags[aws.ToString(tag.TagKey)] = aws.ToString(tag.TagValue)
					}
					resource["tags"] = tags
				}

				// Computed fields
				resource["is_symmetric"] = metadata.KeySpec == types.KeySpecSymmetricDefault
				resource["is_asymmetric"] = metadata.KeySpec != types.KeySpecSymmetricDefault
				resource["is_signing_key"] = metadata.KeyUsage == types.KeyUsageTypeSignVerify
				resource["is_encryption_key"] = metadata.KeyUsage == types.KeyUsageTypeEncryptDecrypt

				// State checks
				resource["is_pending_deletion"] = metadata.KeyState == types.KeyStatePendingDeletion
				resource["is_pending_import"] = metadata.KeyState == types.KeyStatePendingImport
				resource["is_disabled"] = metadata.KeyState == types.KeyStateDisabled

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// analyzeKMSPolicy checks for public access and cross-account access in KMS key policy.
func analyzeKMSPolicy(policy map[string]any, accountID string) (hasPublicAccess, allowsCrossAccount bool) {
	statements, ok := policy["Statement"].([]any)
	if !ok {
		return false, false
	}

	for _, stmt := range statements {
		statement, ok := stmt.(map[string]any)
		if !ok {
			continue
		}

		// Only check Allow statements
		effect, _ := statement["Effect"].(string) //nolint:errcheck // zero value is intentional default
		if effect != policyEffectAllow {
			continue
		}

		principal := statement["Principal"]

		// Check for * principal (public access)
		if principal == "*" {
			// Check for conditions
			if statement["Condition"] == nil {
				hasPublicAccess = true
			}
		}

		// Check for cross-account access
		if principalMap, ok := principal.(map[string]any); ok {
			if awsPrincipal, ok := principalMap["AWS"]; ok {
				checkArn := func(arn string) {
					if arn == "*" {
						hasPublicAccess = true
					} else if strings.HasPrefix(arn, "arn:aws:iam::") {
						parts := strings.Split(arn, ":")
						if len(parts) >= 5 && parts[4] != accountID {
							allowsCrossAccount = true
						}
					}
				}

				switch v := awsPrincipal.(type) {
				case string:
					checkArn(v)
				case []any:
					for _, a := range v {
						if arnStr, ok := a.(string); ok {
							checkArn(arnStr)
						}
					}
				}
			}
		}
	}

	return hasPublicAccess, allowsCrossAccount
}
