package aws

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ptr"
)

// S3BucketResource fetches S3 buckets.
type S3BucketResource struct {
	conn   *Connection
	client S3API // optional, for testing
}

// Name returns the resource type name.
func (r *S3BucketResource) Name() string {
	return "aws_s3_bucket"
}

// getClient returns the S3 client, using the injected mock if available.
func (r *S3BucketResource) getClient() S3API {
	if r.client != nil {
		return r.client
	}
	return s3.NewFromConfig(r.conn.cfg)
}

// Fetch retrieves all S3 buckets with their security configurations.
func (r *S3BucketResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	client := r.getClient()

	// List all buckets (global operation)
	bucketsResp, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, core.WrapError("aws", "s3_bucket", "list", err)
	}

	resources := make([]core.Resource, 0, len(bucketsResp.Buckets))

	for _, bucket := range bucketsResp.Buckets {
		bucketName := aws.ToString(bucket.Name)
		resource := make(core.Resource)
		resource["id"] = bucketName
		resource["name"] = bucketName
		resource["arn"] = "arn:aws:s3:::" + bucketName

		if bucket.CreationDate != nil {
			resource["created_at"] = bucket.CreationDate.String()
		}

		// Get bucket location
		locationResp, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		})
		if err == nil {
			region := string(locationResp.LocationConstraint)
			if region == "" {
				region = "us-east-1" // Default region
			}
			resource["region"] = region
		}

		// Get versioning status
		versioningResp, err := client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
			Bucket: bucket.Name,
		})
		if err == nil {
			resource["versioning_status"] = string(versioningResp.Status)
			resource["has_versioning"] = versioningResp.Status == types.BucketVersioningStatusEnabled
			resource["mfa_delete_enabled"] = versioningResp.MFADelete == types.MFADeleteStatusEnabled
		} else {
			resource["has_versioning"] = false
		}

		// Get encryption configuration
		encryptionResp, err := client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{
			Bucket: bucket.Name,
		})
		if err == nil && encryptionResp.ServerSideEncryptionConfiguration != nil {
			resource["has_encryption"] = true
			for _, rule := range encryptionResp.ServerSideEncryptionConfiguration.Rules {
				if rule.ApplyServerSideEncryptionByDefault != nil {
					resource["encryption_algorithm"] = string(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm)
					resource["kms_key_id"] = aws.ToString(rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID)
					resource["has_kms_encryption"] = rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm == types.ServerSideEncryptionAwsKms
				}
				resource["bucket_key_enabled"] = ptr.Deref(rule.BucketKeyEnabled, false)
			}
		} else {
			resource["has_encryption"] = false
		}

		// Get logging configuration
		loggingResp, err := client.GetBucketLogging(ctx, &s3.GetBucketLoggingInput{
			Bucket: bucket.Name,
		})
		if err == nil && loggingResp.LoggingEnabled != nil {
			resource["has_logging"] = true
			resource["logging_target_bucket"] = aws.ToString(loggingResp.LoggingEnabled.TargetBucket)
			resource["logging_target_prefix"] = aws.ToString(loggingResp.LoggingEnabled.TargetPrefix)
		} else {
			resource["has_logging"] = false
		}

		// Get public access block
		pabResp, err := client.GetPublicAccessBlock(ctx, &s3.GetPublicAccessBlockInput{
			Bucket: bucket.Name,
		})
		if err == nil && pabResp.PublicAccessBlockConfiguration != nil {
			pab := pabResp.PublicAccessBlockConfiguration
			blockPublicAcls := ptr.Deref(pab.BlockPublicAcls, false)
			ignorePublicAcls := ptr.Deref(pab.IgnorePublicAcls, false)
			blockPublicPolicy := ptr.Deref(pab.BlockPublicPolicy, false)
			restrictPublicBuckets := ptr.Deref(pab.RestrictPublicBuckets, false)

			resource["block_public_acls"] = blockPublicAcls
			resource["ignore_public_acls"] = ignorePublicAcls
			resource["block_public_policy"] = blockPublicPolicy
			resource["restrict_public_buckets"] = restrictPublicBuckets

			// All blocks enabled
			resource["public_access_block_enabled"] = blockPublicAcls && ignorePublicAcls && blockPublicPolicy && restrictPublicBuckets
		} else {
			resource["public_access_block_enabled"] = false
			resource["block_public_acls"] = false
			resource["ignore_public_acls"] = false
			resource["block_public_policy"] = false
			resource["restrict_public_buckets"] = false
		}

		// Get bucket ACL
		aclResp, err := client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
			Bucket: bucket.Name,
		})
		if err == nil {
			var hasPublicACL bool
			for _, grant := range aclResp.Grants {
				if grant.Grantee != nil && grant.Grantee.URI != nil {
					uri := aws.ToString(grant.Grantee.URI)
					if strings.Contains(uri, "AllUsers") || strings.Contains(uri, "AuthenticatedUsers") {
						hasPublicACL = true
						break
					}
				}
			}
			resource["has_public_acl"] = hasPublicACL
		} else {
			resource["has_public_acl"] = false
		}

		// Get bucket policy
		policyResp, err := client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{
			Bucket: bucket.Name,
		})
		if err == nil && policyResp.Policy != nil {
			policyDoc := aws.ToString(policyResp.Policy)
			resource["bucket_policy"] = policyDoc
			resource["has_bucket_policy"] = true

			// Analyze policy for public access and secure transport
			var policy map[string]any
			if err := json.Unmarshal([]byte(policyDoc), &policy); err == nil {
				resource["has_public_policy"] = checkPublicBucketPolicy(policy)
				resource["requires_secure_transport"] = checkSecureTransportPolicy(policy)
			}
		} else {
			resource["has_bucket_policy"] = false
			resource["has_public_policy"] = false
			resource["requires_secure_transport"] = false
		}

		// Get lifecycle rules
		lifecycleResp, err := client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
			Bucket: bucket.Name,
		})
		if err == nil && len(lifecycleResp.Rules) > 0 {
			resource["has_lifecycle_rules"] = true
			resource["lifecycle_rule_count"] = len(lifecycleResp.Rules)
		} else {
			resource["has_lifecycle_rules"] = false
			resource["lifecycle_rule_count"] = 0
		}

		// Get object lock configuration
		objectLockResp, err := client.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{
			Bucket: bucket.Name,
		})
		if err == nil && objectLockResp.ObjectLockConfiguration != nil {
			resource["object_lock_enabled"] = objectLockResp.ObjectLockConfiguration.ObjectLockEnabled == types.ObjectLockEnabledEnabled
			if objectLockResp.ObjectLockConfiguration.Rule != nil && objectLockResp.ObjectLockConfiguration.Rule.DefaultRetention != nil {
				resource["object_lock_mode"] = string(objectLockResp.ObjectLockConfiguration.Rule.DefaultRetention.Mode)
				resource["object_lock_days"] = objectLockResp.ObjectLockConfiguration.Rule.DefaultRetention.Days
				resource["object_lock_years"] = objectLockResp.ObjectLockConfiguration.Rule.DefaultRetention.Years
			}
		} else {
			resource["object_lock_enabled"] = false
		}

		// Computed: Overall public status
		hasPublicACL, _ := resource["has_public_acl"].(bool)            //nolint:errcheck // zero value is intentional default
		hasPublicPolicy, _ := resource["has_public_policy"].(bool)      //nolint:errcheck // zero value is intentional default
		pabEnabled, _ := resource["public_access_block_enabled"].(bool) //nolint:errcheck // zero value is intentional default
		resource["is_public"] = (hasPublicACL || hasPublicPolicy) && !pabEnabled

		// Computed: Allows HTTP
		requiresSSL, _ := resource["requires_secure_transport"].(bool) //nolint:errcheck // zero value is intentional default
		resource["allows_http"] = !requiresSSL

		resources = append(resources, resource)
	}

	return resources, nil
}

// checkPublicBucketPolicy checks if a bucket policy allows public access.
func checkPublicBucketPolicy(policy map[string]any) bool {
	statements, ok := policy["Statement"].([]any)
	if !ok {
		return false
	}

	for _, stmt := range statements {
		statement, ok := stmt.(map[string]any)
		if !ok {
			continue
		}

		// Only check Allow statements
		if effect, ok := statement["Effect"].(string); ok && effect != policyEffectAllow {
			continue
		}

		// Check for public principal
		principal := statement["Principal"]
		if principal == "*" {
			// Check for conditions that might restrict access
			if statement["Condition"] == nil {
				return true
			}
		}
	}

	return false
}

// checkSecureTransportPolicy checks if a bucket policy requires HTTPS.
func checkSecureTransportPolicy(policy map[string]any) bool {
	statements, ok := policy["Statement"].([]any)
	if !ok {
		return false
	}

	for _, stmt := range statements {
		statement, ok := stmt.(map[string]any)
		if !ok {
			continue
		}

		// Look for Deny statements with SecureTransport condition
		effect, _ := statement["Effect"].(string) //nolint:errcheck // zero value is intentional default
		if effect != "Deny" {
			continue
		}

		// Check for aws:SecureTransport condition
		condition, ok := statement["Condition"].(map[string]any)
		if !ok {
			continue
		}

		// Check Bool condition
		if boolCond, ok := condition["Bool"].(map[string]any); ok {
			if secureTransport, ok := boolCond["aws:SecureTransport"]; ok {
				if st, ok := secureTransport.(string); ok && st == "false" {
					return true
				}
			}
		}
	}

	return false
}
