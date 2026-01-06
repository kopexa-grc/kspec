package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"

	"github.com/kopexa-grc/kspec/core"
)

// ACMCertificateResource fetches ACM certificates.
type ACMCertificateResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *ACMCertificateResource) Name() string {
	return "aws_acm_certificate"
}

// Fetch retrieves all ACM certificates across configured regions.
func (r *ACMCertificateResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := acm.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := acm.NewListCertificatesPaginator(client, &acm.ListCertificatesInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, certSummary := range page.CertificateSummaryList {
				// Get certificate details
				certResp, err := client.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
					CertificateArn: certSummary.CertificateArn,
				})
				if err != nil {
					continue
				}

				cert := certResp.Certificate
				resource := make(core.Resource)
				resource["id"] = aws.ToString(cert.CertificateArn)
				resource["arn"] = aws.ToString(cert.CertificateArn)
				resource["domain_name"] = aws.ToString(cert.DomainName)
				resource["region"] = region

				resource["status"] = string(cert.Status)
				resource["type"] = string(cert.Type)
				resource["key_algorithm"] = string(cert.KeyAlgorithm)
				resource["signature_algorithm"] = aws.ToString(cert.SignatureAlgorithm)
				resource["serial"] = aws.ToString(cert.Serial)
				resource["subject"] = aws.ToString(cert.Subject)
				resource["issuer"] = aws.ToString(cert.Issuer)

				// Domain validation
				resource["subject_alternative_names"] = cert.SubjectAlternativeNames
				resource["san_count"] = len(cert.SubjectAlternativeNames)

				// Validation method
				if len(cert.DomainValidationOptions) > 0 {
					resource["validation_method"] = string(cert.DomainValidationOptions[0].ValidationMethod)
					resource["is_dns_validated"] = cert.DomainValidationOptions[0].ValidationMethod == types.ValidationMethodDns
					resource["is_email_validated"] = cert.DomainValidationOptions[0].ValidationMethod == types.ValidationMethodEmail
				}

				// Dates
				if cert.NotBefore != nil {
					resource["not_before"] = cert.NotBefore.String()
				}
				if cert.NotAfter != nil {
					resource["not_after"] = cert.NotAfter.String()
					resource["expires_at"] = cert.NotAfter.String()

					// Calculate days until expiration
					daysUntilExpiration := int(time.Until(*cert.NotAfter).Hours() / 24)
					resource["days_until_expiration"] = daysUntilExpiration
					resource["expires_in_30_days"] = daysUntilExpiration <= 30
					resource["expires_in_60_days"] = daysUntilExpiration <= 60
					resource["expires_in_90_days"] = daysUntilExpiration <= 90
					resource["is_expired"] = daysUntilExpiration < 0
				}

				if cert.CreatedAt != nil {
					resource["created_at"] = cert.CreatedAt.String()
				}
				if cert.IssuedAt != nil {
					resource["issued_at"] = cert.IssuedAt.String()
				}
				if cert.ImportedAt != nil {
					resource["imported_at"] = cert.ImportedAt.String()
				}

				// Renewal
				resource["renewal_eligibility"] = string(cert.RenewalEligibility)
				resource["is_eligible_for_renewal"] = cert.RenewalEligibility == types.RenewalEligibilityEligible
				resource["is_ineligible_for_renewal"] = cert.RenewalEligibility == types.RenewalEligibilityIneligible

				if cert.RenewalSummary != nil {
					resource["renewal_status"] = string(cert.RenewalSummary.RenewalStatus)
					if cert.RenewalSummary.UpdatedAt != nil {
						resource["renewal_updated_at"] = cert.RenewalSummary.UpdatedAt.String()
					}
				}

				// In-use
				resource["in_use_by"] = cert.InUseBy
				resource["in_use_count"] = len(cert.InUseBy)
				resource["is_in_use"] = len(cert.InUseBy) > 0
				resource["is_unused"] = len(cert.InUseBy) == 0

				// Failure reason
				if cert.FailureReason != "" {
					resource["failure_reason"] = string(cert.FailureReason)
					resource["has_failed"] = true
				} else {
					resource["has_failed"] = false
				}

				// Revocation
				if cert.RevocationReason != "" {
					resource["revocation_reason"] = string(cert.RevocationReason)
					resource["is_revoked"] = true
				} else {
					resource["is_revoked"] = false
				}
				if cert.RevokedAt != nil {
					resource["revoked_at"] = cert.RevokedAt.String()
				}

				// Extended key usage
				var keyUsages []string
				for _, ku := range cert.ExtendedKeyUsages {
					keyUsages = append(keyUsages, string(ku.Name))
				}
				resource["extended_key_usages"] = keyUsages

				// Key usages
				var usages []string
				for _, ku := range cert.KeyUsages {
					usages = append(usages, string(ku.Name))
				}
				resource["key_usages"] = usages

				// Certificate transparency logging
				if cert.Options != nil {
					resource["certificate_transparency_logging_preference"] = string(cert.Options.CertificateTransparencyLoggingPreference)
					resource["has_ct_logging"] = cert.Options.CertificateTransparencyLoggingPreference == types.CertificateTransparencyLoggingPreferenceEnabled
				}

				// Tags
				tagsResp, err := client.ListTagsForCertificate(ctx, &acm.ListTagsForCertificateInput{
					CertificateArn: cert.CertificateArn,
				})
				if err == nil {
					tags := make(map[string]string)
					for _, tag := range tagsResp.Tags {
						tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
					}
					resource["tags"] = tags
				}

				// Computed
				resource["is_issued"] = cert.Status == types.CertificateStatusIssued
				resource["is_pending_validation"] = cert.Status == types.CertificateStatusPendingValidation
				resource["is_inactive"] = cert.Status == types.CertificateStatusInactive
				resource["is_amazon_issued"] = cert.Type == types.CertificateTypeAmazonIssued
				resource["is_imported"] = cert.Type == types.CertificateTypeImported
				resource["is_private"] = cert.Type == types.CertificateTypePrivate

				// Security checks
				isIssued := cert.Status == types.CertificateStatusIssued
				isExpiresSoon, _ := resource["expires_in_30_days"].(bool) //nolint:errcheck // zero value ok
				resource["needs_attention"] = !isIssued || isExpiresSoon

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
