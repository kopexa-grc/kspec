package network

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/juliankoehn/kspec/core"
)

func TestCertificatesFetch(t *testing.T) {
	// 1. Generate a self-signed cert for testing
	certPEM, keyPEM, err := generateSelfSignedCert()
	if err != nil {
		t.Fatalf("Failed to generate cert: %v", err)
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("Failed to load key pair: %v", err)
	}

	// 2. Start a TLS server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	serverConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	// Just accept connection and close
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			tlsConn := tls.Server(conn, serverConfig)
			tlsConn.Handshake()
			tlsConn.Close()
		}
	}()
	defer listener.Close()

	addr := listener.Addr().String()

	// 3. Fetch
	r := &CertificatesResource{}
	// asset needs domain (hostname) and target (ip:port)
	// For testing, we can use "localhost" or specific IP?
	// Our Fetch logic uses "domain" for SNI and "target" for connection if specified.

	asset := core.Asset{
		FQDN: "example.com",
		Config: map[string]string{
			"target": addr,
			"domain": "example.com",
		},
	}

	resources, err := r.Fetch(context.Background(), asset)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(resources) == 0 {
		t.Fatal("Expected at least one resource (the certificate)")
	}

	res := resources[0]
	if res["isCA"] != true {
		// Our generator makes it a CA
		t.Error("Expected isCA to be true")
	}
	if sub, ok := res["subject"].(map[string]interface{}); ok {
		if sub["commonName"] != "Test Cert" {
			t.Errorf("Expected CommonName 'Test Cert', got %v", sub["commonName"])
		}
	} else {
		t.Error("Subject is not a map")
	}
}

func generateSelfSignedCert() ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test Cert",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certPEM, keyPEM, nil
}
