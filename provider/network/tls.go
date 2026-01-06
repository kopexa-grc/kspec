package network

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/kopexa-grc/kspec/core"
)

// TLSResource provides access to TLS configuration details including supported versions and ciphers.
type TLSResource struct{}

// Name returns the resource type name.
func (r *TLSResource) Name() string {
	return "tls"
}

// Fetch retrieves TLS configuration from the target host specified in the asset.
func (r *TLSResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	domainName := asset.FQDN
	if domainName == "" {
		domainName = config["domain"]
	}
	target := config["target"] // Allow target override

	if domainName == "" && target == "" {
		return nil, fmt.Errorf("missing 'domain' or FQDN in asset for tls resource")
	}
	if domainName == "" {
		// Attempt to extract domain from target if clean
		domainName = target
		if h, _, err := net.SplitHostPort(target); err == nil {
			domainName = h
		}
	} else if target == "" {
		target = domainName
	}

	// Ensure port
	if !strings.Contains(target, ":") {
		target = net.JoinHostPort(target, "443")
	}

	// versions probing
	versionsToProbe := []uint16{
		tls.VersionTLS10,
		tls.VersionTLS11,
		tls.VersionTLS12,
		tls.VersionTLS13,
	}

	supportedVersions := []string{}
	supportedCiphers := make(map[string]bool)
	var peerCertificates []*x509CertWrapper

	// Track connection success to determine if "tls" is effectively "enabled"
	connectedAtLeastOnce := false
	var finalConnState *tls.ConnectionState

	dial := func(version uint16) (*tls.ConnectionState, error) {
		conf := &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // Intentional: security scanner needs to inspect certs even if invalid
			MinVersion:         version,
			MaxVersion:         version,
			ServerName:         domainName, // SNI
		}
		dialer := &tls.Dialer{
			NetDialer: &net.Dialer{
				Timeout: 5 * time.Second,
			},
			Config: conf,
		}
		conn, err := dialer.DialContext(ctx, "tcp", target)
		if err != nil {
			return nil, err
		}
		defer func() { _ = conn.Close() }() //nolint:errcheck // Intentional: Close error is not actionable here
		tlsConn, ok := conn.(*tls.Conn)
		if !ok {
			return nil, fmt.Errorf("connection is not a TLS connection")
		}
		state := tlsConn.ConnectionState()
		return &state, nil
	}

	for _, v := range versionsToProbe {
		state, err := dial(v)
		if err == nil {
			connectedAtLeastOnce = true
			verStr := versionToString(v)
			supportedVersions = append(supportedVersions, verStr)
			cipherStr := tls.CipherSuiteName(state.CipherSuite)
			supportedCiphers[cipherStr] = true

			// We prioritize the state from the highest supported version for certificate info
			finalConnState = state
		}
	}

	if !connectedAtLeastOnce {
		return nil, fmt.Errorf("failed to connect to %s with any TLS version", target)
	}

	// Prepare ciphers list
	ciphers := []string{}
	for c := range supportedCiphers {
		ciphers = append(ciphers, c)
	}

	// Process Certificates
	if finalConnState != nil {
		certs := finalConnState.PeerCertificates
		if len(certs) > 0 {
			// Verification Logic
			isVerified := false
			intermediates := x509.NewCertPool()
			for i := 1; i < len(certs); i++ {
				intermediates.AddCert(certs[i])
			}
			// Verify against system roots
			opts := x509.VerifyOptions{
				DNSName:       domainName,
				Intermediates: intermediates,
			}
			if _, err := certs[0].Verify(opts); err == nil {
				isVerified = true
			}

			for i, cert := range certs {
				daysUntil := int(time.Until(cert.NotAfter).Hours() / 24)
				validityDays := int(cert.NotAfter.Sub(cert.NotBefore).Hours() / 24)

				// For the leaf cert (index 0), we attach the chain verification result
				certVerified := false
				if i == 0 {
					certVerified = isVerified
				}

				certMap := &x509CertWrapper{
					Subject:            subjectWrapper{CommonName: cert.Subject.CommonName},
					Issuer:             issuerWrapper{CommonName: cert.Issuer.CommonName},
					NotBefore:          cert.NotBefore,
					NotAfter:           cert.NotAfter,
					DNSNames:           cert.DNSNames,
					IsCA:               cert.IsCA,
					SignatureAlgorithm: cert.SignatureAlgorithm.String(),
					ExpiresIn:          expiresInWrapper{Days: daysUntil},
					ValidityDays:       validityDays,
					IsVerified:         certVerified,
				}
				peerCertificates = append(peerCertificates, certMap)
			}
		}
	}

	certsList := []map[string]interface{}{}
	for _, c := range peerCertificates {
		certsList = append(certsList, map[string]interface{}{
			"subject": map[string]interface{}{
				"commonName": c.Subject.CommonName,
			},
			"issuer": map[string]interface{}{
				"commonName": c.Issuer.CommonName,
			},
			"notBefore":        c.NotBefore,
			"notAfter":         c.NotAfter,
			"dnsNames":         c.DNSNames,
			"isCA":             c.IsCA,
			"signingAlgorithm": c.SignatureAlgorithm,
			"expiresIn": map[string]interface{}{
				"days": c.ExpiresIn.Days,
			},
			"validityDays": c.ValidityDays,
			"isVerified":   c.IsVerified,
		})
	}

	res := core.Resource{
		"name":         domainName,
		"versions":     supportedVersions,
		"ciphers":      ciphers,
		"certificates": certsList,
		"params":       map[string]interface{}{},
	}

	return []core.Resource{res}, nil
}

func versionToString(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "tls1.0"
	case tls.VersionTLS11:
		return "tls1.1"
	case tls.VersionTLS12:
		return "tls1.2"
	case tls.VersionTLS13:
		return "tls1.3"
	case 0x0300: // SSLv3 (deprecated constant)
		return "ssl3.0"
	default:
		return fmt.Sprintf("unknown_%x", v)
	}
}

type x509CertWrapper struct {
	Subject            subjectWrapper
	Issuer             issuerWrapper
	NotBefore          time.Time
	NotAfter           time.Time
	DNSNames           []string
	IsCA               bool
	SignatureAlgorithm string
	ExpiresIn          expiresInWrapper
	ValidityDays       int
	IsVerified         bool
}

type subjectWrapper struct {
	CommonName string
}
type issuerWrapper struct {
	CommonName string
}
type expiresInWrapper struct {
	Days int
}
