// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"

	"github.com/kopexa-grc/kspec/core"
)

// ECRClient is the interface for ECR operations.
//
//nolint:dupl // AWS SDK client interfaces have similar structures by design
type ECRClient interface {
	DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error)
	GetRepositoryPolicy(ctx context.Context, params *ecr.GetRepositoryPolicyInput, optFns ...func(*ecr.Options)) (*ecr.GetRepositoryPolicyOutput, error)
	GetLifecyclePolicy(ctx context.Context, params *ecr.GetLifecyclePolicyInput, optFns ...func(*ecr.Options)) (*ecr.GetLifecyclePolicyOutput, error)
	DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error)
	ListTagsForResource(ctx context.Context, params *ecr.ListTagsForResourceInput, optFns ...func(*ecr.Options)) (*ecr.ListTagsForResourceOutput, error)
}

// ECRClientFactory creates ECR clients for specific regions.
type ECRClientFactory func(region string) ECRClient

// ECRRepository fetches ECR repositories.
type ECRRepository struct {
	clientFactory ECRClientFactory
	accountID     string
	regions       []string
}

// NewECRRepository creates a new ECRRepository resource.
func NewECRRepository(clientFactory ECRClientFactory, accountID string, regions []string) *ECRRepository {
	return &ECRRepository{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *ECRRepository) Name() string {
	return "aws_ecr_repository"
}

// Fetch retrieves all ECR repositories across configured regions.
func (r *ECRRepository) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := ecr.NewDescribeRepositoriesPaginator(client, &ecr.DescribeRepositoriesInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, repo := range page.Repositories {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(repo.RepositoryName)
				resource["arn"] = aws.ToString(repo.RepositoryArn)
				resource["name"] = aws.ToString(repo.RepositoryName)
				resource["region"] = region

				resource["registry_id"] = aws.ToString(repo.RegistryId)
				resource["repository_uri"] = aws.ToString(repo.RepositoryUri)

				// Image tag mutability
				resource["image_tag_mutability"] = string(repo.ImageTagMutability)
				resource["is_immutable"] = repo.ImageTagMutability == types.ImageTagMutabilityImmutable
				resource["allows_tag_overwrite"] = repo.ImageTagMutability == types.ImageTagMutabilityMutable

				// Encryption
				if repo.EncryptionConfiguration != nil {
					resource["encryption_type"] = string(repo.EncryptionConfiguration.EncryptionType)
					resource["kms_key"] = aws.ToString(repo.EncryptionConfiguration.KmsKey)
					resource["has_encryption"] = true
					resource["has_cmk_encryption"] = repo.EncryptionConfiguration.EncryptionType == types.EncryptionTypeKms && aws.ToString(repo.EncryptionConfiguration.KmsKey) != ""
				} else {
					resource["has_encryption"] = false
					resource["has_cmk_encryption"] = false
				}

				// Image scanning
				if repo.ImageScanningConfiguration != nil {
					resource["scan_on_push"] = repo.ImageScanningConfiguration.ScanOnPush
					resource["has_scan_on_push"] = repo.ImageScanningConfiguration.ScanOnPush
				} else {
					resource["scan_on_push"] = false
					resource["has_scan_on_push"] = false
				}

				// Created at
				if repo.CreatedAt != nil {
					resource["created_at"] = repo.CreatedAt.String()
				}

				// Get repository policy
				policyResp, err := client.GetRepositoryPolicy(ctx, &ecr.GetRepositoryPolicyInput{
					RepositoryName: repo.RepositoryName,
				})
				if err == nil {
					resource["has_policy"] = true
					resource["policy"] = aws.ToString(policyResp.PolicyText)

					// Check for public access
					isPublic := checkPublicECRPolicy(aws.ToString(policyResp.PolicyText))
					resource["is_public"] = isPublic
					resource["allows_public_access"] = isPublic
				} else {
					resource["has_policy"] = false
					resource["is_public"] = false
				}

				// Get lifecycle policy
				lifecycleResp, err := client.GetLifecyclePolicy(ctx, &ecr.GetLifecyclePolicyInput{
					RepositoryName: repo.RepositoryName,
				})
				if err == nil {
					resource["has_lifecycle_policy"] = true
					resource["lifecycle_policy"] = aws.ToString(lifecycleResp.LifecyclePolicyText)
				} else {
					resource["has_lifecycle_policy"] = false
				}

				// Get image count
				imagesResp, err := client.DescribeImages(ctx, &ecr.DescribeImagesInput{
					RepositoryName: repo.RepositoryName,
					MaxResults:     aws.Int32(1000),
				})
				if err == nil {
					resource["image_count"] = len(imagesResp.ImageDetails)
					resource["has_images"] = len(imagesResp.ImageDetails) > 0

					// Count images with vulnerabilities
					vulnerableCount := 0
					criticalCount := 0
					highCount := 0
					for _, img := range imagesResp.ImageDetails {
						if img.ImageScanStatus != nil && img.ImageScanStatus.Status == types.ScanStatusComplete {
							if img.ImageScanFindingsSummary != nil {
								if counts := img.ImageScanFindingsSummary.FindingSeverityCounts; counts != nil {
									if c, ok := counts[string(types.FindingSeverityCritical)]; ok && c > 0 {
										criticalCount++
									}
									if c, ok := counts[string(types.FindingSeverityHigh)]; ok && c > 0 {
										highCount++
									}
								}
							}
						}
					}
					resource["images_with_critical_vulns"] = criticalCount
					resource["images_with_high_vulns"] = highCount
					resource["vulnerable_image_count"] = vulnerableCount
					resource["has_vulnerable_images"] = criticalCount > 0 || highCount > 0
				} else {
					resource["image_count"] = 0
					resource["has_images"] = false
				}

				// Get tags
				tagsResp, err := client.ListTagsForResource(ctx, &ecr.ListTagsForResourceInput{
					ResourceArn: repo.RepositoryArn,
				})
				if err == nil {
					tags := make(map[string]string)
					for _, tag := range tagsResp.Tags {
						tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
					}
					resource["tags"] = tags
				}

				// Computed security baseline
				isImmutable, _ := resource["is_immutable"].(bool)          //nolint:errcheck // zero value ok
				hasScan, _ := resource["has_scan_on_push"].(bool)          //nolint:errcheck // zero value ok
				hasLifecycle, _ := resource["has_lifecycle_policy"].(bool) //nolint:errcheck // zero value ok
				resource["meets_security_baseline"] = isImmutable && hasScan && hasLifecycle

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

func checkPublicECRPolicy(policyJSON string) bool {
	var policy struct {
		Statement []struct {
			Effect    string      `json:"Effect"`
			Principal interface{} `json:"Principal"`
		} `json:"Statement"`
	}

	if err := json.Unmarshal([]byte(policyJSON), &policy); err != nil {
		return false
	}

	for _, stmt := range policy.Statement {
		if stmt.Effect != policyEffectAllow {
			continue
		}

		switch p := stmt.Principal.(type) {
		case string:
			if p == "*" {
				return true
			}
		case map[string]interface{}:
			if awsPrincipal, ok := p["AWS"]; ok {
				switch v := awsPrincipal.(type) {
				case string:
					if v == "*" {
						return true
					}
				case []interface{}:
					for _, item := range v {
						if s, ok := item.(string); ok && s == "*" {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// ECRImage fetches ECR images.
type ECRImage struct {
	clientFactory ECRClientFactory
	accountID     string
	regions       []string
}

// NewECRImage creates a new ECRImage resource.
func NewECRImage(clientFactory ECRClientFactory, accountID string, regions []string) *ECRImage {
	return &ECRImage{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *ECRImage) Name() string {
	return "aws_ecr_image"
}

// Fetch retrieves ECR images across configured regions.
func (r *ECRImage) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		// First get all repositories
		repoPaginator := ecr.NewDescribeRepositoriesPaginator(client, &ecr.DescribeRepositoriesInput{})

		for repoPaginator.HasMorePages() {
			repoPage, err := repoPaginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, repo := range repoPage.Repositories {
				// Get images for this repository
				imgPaginator := ecr.NewDescribeImagesPaginator(client, &ecr.DescribeImagesInput{
					RepositoryName: repo.RepositoryName,
				})

				for imgPaginator.HasMorePages() {
					imgPage, err := imgPaginator.NextPage(ctx)
					if err != nil {
						continue
					}

					for _, img := range imgPage.ImageDetails {
						resource := make(core.Resource)
						digest := aws.ToString(img.ImageDigest)
						resource["id"] = aws.ToString(repo.RepositoryName) + ":" + digest
						resource["digest"] = digest
						resource["repository_name"] = aws.ToString(repo.RepositoryName)
						resource["registry_id"] = aws.ToString(img.RegistryId)
						resource["region"] = region

						// Tags
						resource["image_tags"] = img.ImageTags
						resource["tag_count"] = len(img.ImageTags)
						resource["is_tagged"] = len(img.ImageTags) > 0
						resource["is_untagged"] = len(img.ImageTags) == 0

						// Size
						if img.ImageSizeInBytes != nil {
							resource["size_bytes"] = *img.ImageSizeInBytes
							resource["size_mb"] = *img.ImageSizeInBytes / (1024 * 1024)
						}

						// Pushed at
						if img.ImagePushedAt != nil {
							resource["pushed_at"] = img.ImagePushedAt.String()
						}

						// Manifest media type
						resource["manifest_media_type"] = aws.ToString(img.ImageManifestMediaType)
						resource["artifact_media_type"] = aws.ToString(img.ArtifactMediaType)

						// Scan status
						if img.ImageScanStatus != nil {
							resource["scan_status"] = string(img.ImageScanStatus.Status)
							resource["is_scan_complete"] = img.ImageScanStatus.Status == types.ScanStatusComplete
							resource["is_scan_in_progress"] = img.ImageScanStatus.Status == types.ScanStatusInProgress
							resource["is_scan_failed"] = img.ImageScanStatus.Status == types.ScanStatusFailed
							resource["has_been_scanned"] = img.ImageScanStatus.Status == types.ScanStatusComplete
						} else {
							resource["has_been_scanned"] = false
						}

						// Scan findings
						if img.ImageScanFindingsSummary != nil {
							summary := img.ImageScanFindingsSummary

							if summary.ImageScanCompletedAt != nil {
								resource["scan_completed_at"] = summary.ImageScanCompletedAt.String()
							}
							if summary.VulnerabilitySourceUpdatedAt != nil {
								resource["vuln_source_updated_at"] = summary.VulnerabilitySourceUpdatedAt.String()
							}

							// Finding counts by severity
							criticalCount := 0
							highCount := 0
							mediumCount := 0
							lowCount := 0
							informationalCount := 0
							undefinedCount := 0

							if summary.FindingSeverityCounts != nil {
								if c, ok := summary.FindingSeverityCounts[string(types.FindingSeverityCritical)]; ok {
									criticalCount = int(c)
								}
								if c, ok := summary.FindingSeverityCounts[string(types.FindingSeverityHigh)]; ok {
									highCount = int(c)
								}
								if c, ok := summary.FindingSeverityCounts[string(types.FindingSeverityMedium)]; ok {
									mediumCount = int(c)
								}
								if c, ok := summary.FindingSeverityCounts[string(types.FindingSeverityLow)]; ok {
									lowCount = int(c)
								}
								if c, ok := summary.FindingSeverityCounts[string(types.FindingSeverityInformational)]; ok {
									informationalCount = int(c)
								}
								if c, ok := summary.FindingSeverityCounts[string(types.FindingSeverityUndefined)]; ok {
									undefinedCount = int(c)
								}
							}

							resource["critical_count"] = criticalCount
							resource["high_count"] = highCount
							resource["medium_count"] = mediumCount
							resource["low_count"] = lowCount
							resource["informational_count"] = informationalCount
							resource["undefined_count"] = undefinedCount
							resource["total_finding_count"] = criticalCount + highCount + mediumCount + lowCount + informationalCount + undefinedCount

							resource["has_critical_findings"] = criticalCount > 0
							resource["has_high_findings"] = highCount > 0
							resource["has_vulnerabilities"] = criticalCount > 0 || highCount > 0 || mediumCount > 0
							resource["is_clean"] = criticalCount == 0 && highCount == 0 && mediumCount == 0
						}

						// Last recorded pull
						if img.LastRecordedPullTime != nil {
							resource["last_pull_time"] = img.LastRecordedPullTime.String()
						}

						resources = append(resources, resource)
					}
				}
			}
		}
	}

	return resources, nil
}
