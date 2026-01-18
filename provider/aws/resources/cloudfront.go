// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"

	"github.com/kopexa-grc/kspec/core"
)

// CloudFrontClient is the interface for CloudFront operations.
type CloudFrontClient interface {
	ListDistributions(ctx context.Context, params *cloudfront.ListDistributionsInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListDistributionsOutput, error)
	GetDistribution(ctx context.Context, params *cloudfront.GetDistributionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetDistributionOutput, error)
	ListTagsForResource(ctx context.Context, params *cloudfront.ListTagsForResourceInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListTagsForResourceOutput, error)
}

// CloudFrontDistribution fetches CloudFront distributions.
type CloudFrontDistribution struct {
	client    CloudFrontClient
	accountID string
}

// NewCloudFrontDistribution creates a new CloudFrontDistribution resource.
func NewCloudFrontDistribution(client CloudFrontClient, accountID string) *CloudFrontDistribution {
	return &CloudFrontDistribution{
		client:    client,
		accountID: accountID,
	}
}

// Name returns the resource type name.
func (r *CloudFrontDistribution) Name() string {
	return "aws_cloudfront_distribution"
}

// Fetch retrieves all CloudFront distributions.
func (r *CloudFrontDistribution) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	paginator := cloudfront.NewListDistributionsPaginator(r.client, &cloudfront.ListDistributionsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			continue
		}

		if page.DistributionList == nil {
			continue
		}

		for _, dist := range page.DistributionList.Items {
			resource := make(core.Resource)
			resource["id"] = aws.ToString(dist.Id)
			resource["arn"] = aws.ToString(dist.ARN)
			resource["domain_name"] = aws.ToString(dist.DomainName)
			resource["region"] = "global"

			resource["status"] = aws.ToString(dist.Status)
			resource["enabled"] = dist.Enabled
			resource["is_enabled"] = dist.Enabled
			resource["comment"] = aws.ToString(dist.Comment)
			resource["price_class"] = string(dist.PriceClass)
			resource["http_version"] = string(dist.HttpVersion)
			resource["is_ipv6_enabled"] = dist.IsIPV6Enabled

			// Aliases (CNAMEs)
			if dist.Aliases != nil {
				resource["aliases"] = dist.Aliases.Items
				resource["alias_count"] = dist.Aliases.Quantity
			} else {
				resource["alias_count"] = 0
			}

			// Origins
			if dist.Origins != nil {
				originIDs := make([]string, 0, len(dist.Origins.Items))
				originDomains := make([]string, 0, len(dist.Origins.Items))
				hasS3Origin := false
				hasCustomOrigin := false

				for _, origin := range dist.Origins.Items {
					originIDs = append(originIDs, aws.ToString(origin.Id))
					originDomains = append(originDomains, aws.ToString(origin.DomainName))

					domain := aws.ToString(origin.DomainName)
					if strings.Contains(domain, ".s3.") || strings.HasSuffix(domain, ".s3.amazonaws.com") {
						hasS3Origin = true
					}
					if origin.CustomOriginConfig != nil {
						hasCustomOrigin = true
					}
				}

				resource["origin_ids"] = originIDs
				resource["origin_domains"] = originDomains
				resource["origin_count"] = dist.Origins.Quantity
				resource["has_s3_origin"] = hasS3Origin
				resource["has_custom_origin"] = hasCustomOrigin
			}

			// Default cache behavior
			if dist.DefaultCacheBehavior != nil {
				dcb := dist.DefaultCacheBehavior
				resource["viewer_protocol_policy"] = string(dcb.ViewerProtocolPolicy)
				resource["compress"] = dcb.Compress

				// Check for HTTPS enforcement
				viewerPolicy := string(dcb.ViewerProtocolPolicy)
				resource["enforces_https"] = viewerPolicy == "https-only" || viewerPolicy == "redirect-to-https"
				resource["allows_http"] = viewerPolicy == "allow-all"

				// Lambda@Edge or CloudFront Functions
				if dcb.LambdaFunctionAssociations != nil && dcb.LambdaFunctionAssociations.Quantity != nil && *dcb.LambdaFunctionAssociations.Quantity > 0 {
					resource["has_lambda_edge"] = true
					resource["lambda_edge_count"] = dcb.LambdaFunctionAssociations.Quantity
				} else {
					resource["has_lambda_edge"] = false
				}

				if dcb.FunctionAssociations != nil && dcb.FunctionAssociations.Quantity != nil && *dcb.FunctionAssociations.Quantity > 0 {
					resource["has_cloudfront_functions"] = true
					resource["cloudfront_function_count"] = dcb.FunctionAssociations.Quantity
				} else {
					resource["has_cloudfront_functions"] = false
				}

				// Field-level encryption
				resource["field_level_encryption_id"] = aws.ToString(dcb.FieldLevelEncryptionId)
				resource["has_field_level_encryption"] = dcb.FieldLevelEncryptionId != nil && aws.ToString(dcb.FieldLevelEncryptionId) != ""
			}

			// Cache behaviors count
			if dist.CacheBehaviors != nil {
				resource["cache_behavior_count"] = dist.CacheBehaviors.Quantity
			} else {
				resource["cache_behavior_count"] = 0
			}

			// Custom error responses
			if dist.CustomErrorResponses != nil {
				resource["custom_error_response_count"] = dist.CustomErrorResponses.Quantity
				resource["has_custom_error_responses"] = dist.CustomErrorResponses.Quantity != nil && *dist.CustomErrorResponses.Quantity > 0
			} else {
				resource["has_custom_error_responses"] = false
			}

			// Restrictions (geo)
			if dist.Restrictions != nil && dist.Restrictions.GeoRestriction != nil {
				geo := dist.Restrictions.GeoRestriction
				resource["geo_restriction_type"] = string(geo.RestrictionType)
				resource["geo_restriction_locations"] = geo.Items
				resource["has_geo_restriction"] = string(geo.RestrictionType) != "none"
			} else {
				resource["has_geo_restriction"] = false
			}

			// SSL/TLS certificate
			if dist.ViewerCertificate != nil {
				vc := dist.ViewerCertificate
				resource["ssl_support_method"] = string(vc.SSLSupportMethod)
				resource["minimum_protocol_version"] = string(vc.MinimumProtocolVersion)
				resource["acm_certificate_arn"] = aws.ToString(vc.ACMCertificateArn)
				resource["iam_certificate_id"] = aws.ToString(vc.IAMCertificateId)
				resource["cloudfront_default_certificate"] = vc.CloudFrontDefaultCertificate

				// Check for custom certificate
				resource["has_custom_certificate"] = (vc.ACMCertificateArn != nil && aws.ToString(vc.ACMCertificateArn) != "") ||
					(vc.IAMCertificateId != nil && aws.ToString(vc.IAMCertificateId) != "")

				// Check for modern TLS
				minProtocol := string(vc.MinimumProtocolVersion)
				resource["has_modern_tls"] = strings.Contains(minProtocol, "TLSv1.2") || strings.Contains(minProtocol, "TLSv1.3")
			}

			// Web ACL (WAF)
			resource["web_acl_id"] = aws.ToString(dist.WebACLId)
			resource["has_waf"] = dist.WebACLId != nil && aws.ToString(dist.WebACLId) != ""

			// Logging
			// Need to get full distribution for logging config
			fullDistResp, err := r.client.GetDistribution(ctx, &cloudfront.GetDistributionInput{
				Id: dist.Id,
			})
			if err == nil && fullDistResp.Distribution != nil && fullDistResp.Distribution.DistributionConfig != nil {
				dc := fullDistResp.Distribution.DistributionConfig

				if dc.Logging != nil {
					resource["logging_enabled"] = dc.Logging.Enabled
					resource["has_logging"] = dc.Logging.Enabled
					resource["logging_bucket"] = aws.ToString(dc.Logging.Bucket)
					resource["logging_prefix"] = aws.ToString(dc.Logging.Prefix)
					resource["logging_include_cookies"] = dc.Logging.IncludeCookies
				} else {
					resource["has_logging"] = false
				}

				// Default root object
				resource["default_root_object"] = aws.ToString(dc.DefaultRootObject)
				resource["has_default_root_object"] = dc.DefaultRootObject != nil && aws.ToString(dc.DefaultRootObject) != ""
			}

			// Tags
			tagsResp, err := r.client.ListTagsForResource(ctx, &cloudfront.ListTagsForResourceInput{
				Resource: dist.ARN,
			})
			if err == nil && tagsResp.Tags != nil {
				tags := make(map[string]string)
				for _, tag := range tagsResp.Tags.Items {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags
			}

			// Computed security baseline
			enforcesHTTPS, _ := resource["enforces_https"].(bool) //nolint:errcheck // zero value ok
			hasWAF, _ := resource["has_waf"].(bool)               //nolint:errcheck // zero value ok
			hasLogging, _ := resource["has_logging"].(bool)       //nolint:errcheck // zero value ok
			hasModernTLS, _ := resource["has_modern_tls"].(bool)  //nolint:errcheck // zero value ok
			resource["meets_security_baseline"] = enforcesHTTPS && hasWAF && hasLogging && hasModernTLS

			resources = append(resources, resource)
		}
	}

	return resources, nil
}
