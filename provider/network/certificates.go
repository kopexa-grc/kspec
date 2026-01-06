package network

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

// CertificatesResource provides access to TLS certificate information from remote hosts.
type CertificatesResource struct{}

// Name returns the resource type name.
func (r *CertificatesResource) Name() string {
	return "certificate"
}

// Fetch retrieves TLS certificates from the target host specified in the asset.
func (r *CertificatesResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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

	// Provide all certificates in the chain? Or just leaf?
	// Usually 'certificate' resource might be a collection where we return each cert in the chain as a resource?
	// Or one resource with a list?
	// The snippet suggests "CertificatesToMqlCertificates" taking a collection.
	// We'll return each certificate in the chain as a separate resource,
	// so the user can query "certificate.where(c, c.is_ca)" etc.

	for i, cert := range certs {
		res := parseCertificate(cert)
		// Add metadata about position in chain?
		res["chain_index"] = i
		res["is_leaf"] = (i == 0)
		res["target"] = target
		resources = append(resources, res)
	}

	return resources, nil
}

func parseCertificate(cert *x509.Certificate) core.Resource {
	// Calculate fingerprints
	sha256Fingerprint := hex.EncodeToString(sha256Hash(cert.Raw))

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
	}

	// Extensions (simplified)
	res["keyUsage"] = parseKeyUsage(cert.KeyUsage)
	res["extKeyUsage"] = parseExtKeyUsage(cert.ExtKeyUsage)

	return res
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
