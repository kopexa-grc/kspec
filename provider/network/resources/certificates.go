// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/kopexa-grc/kspec/core"
)

const (
	// Certificate validity constants per CA/Browser Forum Ballot SC-081v3
	// Timeline: 398 days (current) -> 200 days (Mar 2026) -> 100 days (Mar 2027) -> 47 days (Mar 2029)

	// MaxValidityDaysCurrent is the current maximum (until March 2026)
	MaxValidityDaysCurrent = 398
	// MaxValidityDays2026 is the maximum starting March 15, 2026
	MaxValidityDays2026 = 200
	// MaxValidityDays2027 is the maximum starting March 15, 2027
	MaxValidityDays2027 = 100
	// MaxValidityDays2029 is the maximum starting March 15, 2029
	MaxValidityDays2029 = 47

	// MaxValidityDays is set to the current standard
	MaxValidityDays = MaxValidityDaysCurrent

	// ExpirationWarningDays is the number of days before expiration to warn
	ExpirationWarningDays = 30
)

// Certificate provides access to TLS certificate information from remote hosts.
type Certificate struct{}

// NewCertificate creates a new Certificate resource instance.
func NewCertificate() *Certificate {
	return &Certificate{}
}

// Name returns the resource type name.
func (r *Certificate) Name() string {
	return "certificate"
}

// Fetch retrieves TLS certificates from the target host specified in the asset.
func (r *Certificate) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	domainName := asset.FQDN
	if domainName == "" {
		domainName = asset.Config["domain"]
	}
	target := asset.Config["target"]

	if domainName == "" && target == "" {
		return nil, fmt.Errorf("missing 'domain' or FQDN in asset for certificate resource")
	}

	if target == "" {
		target = domainName
	}
	if !strings.Contains(target, ":") {
		target = net.JoinHostPort(target, "443")
	}

	// Dial to get certificates
	conf := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // Intentional: security scanner needs to inspect certs even if invalid
		ServerName:         domainName,
	}
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout: 10 * time.Second,
		},
		Config: conf,
	}
	conn, err := dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", target, err)
	}
	defer func() { _ = conn.Close() }() //nolint:errcheck // Intentional: Close error is not actionable here

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return nil, fmt.Errorf("connection is not a TLS connection")
	}
	state := tlsConn.ConnectionState()
	certs := state.PeerCertificates

	resources := make([]core.Resource, 0, len(certs))

	// Verify certificate chain against system roots
	isChainVerified := false
	if len(certs) > 0 {
		intermediates := x509.NewCertPool()
		for i := 1; i < len(certs); i++ {
			intermediates.AddCert(certs[i])
		}
		opts := x509.VerifyOptions{
			DNSName:       domainName,
			Intermediates: intermediates,
		}
		if _, err := certs[0].Verify(opts); err == nil {
			isChainVerified = true
		}
	}

	for i, cert := range certs {
		res := parseCertificate(cert, domainName)

		// Add metadata about position in chain
		res["chain_index"] = i
		res["is_leaf"] = (i == 0)
		res["target"] = target

		// Add calculated time fields
		daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)
		validityDays := int(cert.NotAfter.Sub(cert.NotBefore).Hours() / 24)

		res["expiresIn"] = map[string]interface{}{
			"days": daysUntilExpiry,
		}
		res["validityDays"] = validityDays
		res["isExpired"] = daysUntilExpiry < 0
		res["isExpiringSoon"] = daysUntilExpiry >= 0 && daysUntilExpiry <= ExpirationWarningDays
		res["hasLongValidity"] = validityDays > MaxValidityDays

		// Only leaf cert gets chain verification result
		if i == 0 {
			res["isVerified"] = isChainVerified
			// Check if self-signed (subject == issuer for single cert or leaf with no proper chain)
			res["isSelfSigned"] = cert.Subject.CommonName == cert.Issuer.CommonName && !isChainVerified
		} else {
			res["isVerified"] = false
			res["isSelfSigned"] = false
		}

		resources = append(resources, res)
	}

	return resources, nil
}

func parseCertificate(cert *x509.Certificate, domainName string) core.Resource {
	// Calculate fingerprints
	sha256Fingerprint := hex.EncodeToString(sha256Hash(cert.Raw))

	// Check if domain name matches certificate
	domainMatches := false
	if domainName != "" {
		if cert.Subject.CommonName == domainName {
			domainMatches = true
		} else {
			for _, dns := range cert.DNSNames {
				if dns == domainName || matchWildcard(dns, domainName) {
					domainMatches = true
					break
				}
			}
		}
	}

	res := core.Resource{
		"serialNumber":       cert.SerialNumber.String(), // BigInt to string
		"subject":            pkixNameMap(cert.Subject),
		"issuer":             pkixNameMap(cert.Issuer),
		"notBefore":          cert.NotBefore,
		"notAfter":           cert.NotAfter,
		"dnsNames":           cert.DNSNames,
		"emailAddresses":     cert.EmailAddresses,
		"ipAddresses":        ipsToStrings(cert.IPAddresses),
		"uris":               urisToStrings(cert.URIs),
		"isCA":               cert.IsCA,
		"signatureAlgorithm": cert.SignatureAlgorithm.String(),
		"publicKeyAlgorithm": cert.PublicKeyAlgorithm.String(),
		"version":            cert.Version,
		"fingerprint_sha256": sha256Fingerprint,
		"domainMatches":      domainMatches,
	}

	// Extensions (simplified)
	res["keyUsage"] = parseKeyUsage(cert.KeyUsage)
	res["extKeyUsage"] = parseExtKeyUsage(cert.ExtKeyUsage)

	return res
}

// matchWildcard checks if a wildcard certificate pattern matches the domain
func matchWildcard(pattern, domain string) bool {
	if !strings.HasPrefix(pattern, "*.") {
		return false
	}
	// *.example.com should match foo.example.com but not foo.bar.example.com
	suffix := pattern[1:] // Remove the *
	if strings.HasSuffix(domain, suffix) {
		prefix := strings.TrimSuffix(domain, suffix)
		// Prefix should not contain any dots (single level only)
		return !strings.Contains(prefix, ".")
	}
	return false
}

func pkixNameMap(name pkix.Name) map[string]interface{} {
	return map[string]interface{}{
		"commonName":         name.CommonName,
		"serialNumber":       name.SerialNumber,
		"country":            name.Country,
		"organization":       name.Organization,
		"organizationalUnit": name.OrganizationalUnit,
		"locality":           name.Locality,
		"province":           name.Province,
		"streetAddress":      name.StreetAddress,
		"postalCode":         name.PostalCode,
	}
}

func sha256Hash(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func parseKeyUsage(ku x509.KeyUsage) []string {
	var usages []string
	if ku&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "DigitalSignature")
	}
	if ku&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "ContentCommitment")
	}
	if ku&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "KeyEncipherment")
	}
	if ku&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "DataEncipherment")
	}
	if ku&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "KeyAgreement")
	}
	if ku&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "CertSign")
	}
	if ku&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "CRLSign")
	}
	if ku&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "EncipherOnly")
	}
	if ku&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "DecipherOnly")
	}
	return usages
}

func parseExtKeyUsage(eku []x509.ExtKeyUsage) []string {
	var usages []string
	for _, u := range eku {
		switch u { //nolint:exhaustive // x509.ExtKeyUsageAny handled in default as unknown
		case x509.ExtKeyUsageServerAuth:
			usages = append(usages, "ServerAuth")
		case x509.ExtKeyUsageClientAuth:
			usages = append(usages, "ClientAuth")
		case x509.ExtKeyUsageCodeSigning:
			usages = append(usages, "CodeSigning")
		case x509.ExtKeyUsageEmailProtection:
			usages = append(usages, "EmailProtection")
		case x509.ExtKeyUsageIPSECEndSystem:
			usages = append(usages, "IPSECEndSystem")
		case x509.ExtKeyUsageIPSECTunnel:
			usages = append(usages, "IPSECTunnel")
		case x509.ExtKeyUsageIPSECUser:
			usages = append(usages, "IPSECUser")
		case x509.ExtKeyUsageTimeStamping:
			usages = append(usages, "TimeStamping")
		case x509.ExtKeyUsageOCSPSigning:
			usages = append(usages, "OCSPSigning")
		case x509.ExtKeyUsageMicrosoftServerGatedCrypto:
			usages = append(usages, "MicrosoftServerGatedCrypto")
		case x509.ExtKeyUsageNetscapeServerGatedCrypto:
			usages = append(usages, "NetscapeServerGatedCrypto")
		case x509.ExtKeyUsageMicrosoftCommercialCodeSigning:
			usages = append(usages, "MicrosoftCommercialCodeSigning")
		case x509.ExtKeyUsageMicrosoftKernelCodeSigning:
			usages = append(usages, "MicrosoftKernelCodeSigning")
		default:
			usages = append(usages, fmt.Sprintf("Unknown(%d)", u))
		}
	}
	return usages
}

func ipsToStrings(ips []net.IP) []string {
	res := make([]string, 0, len(ips))
	for _, ip := range ips {
		res = append(res, ip.String())
	}
	return res
}

func urisToStrings(uris []*url.URL) []string {
	res := make([]string, 0, len(uris))
	for _, u := range uris {
		res = append(res, u.String())
	}
	return res
}
