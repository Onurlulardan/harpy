package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"slices"
	"testing"
	"time"
)

func TestGeneratedRuntimeCertificatesHaveExactSANsAndMatchingKeys(t *testing.T) {
	set, now := generatedSetForTest(t)
	roots := certificatePool(t, set.files["local-ca.crt"])
	tests := []struct {
		name string
		dns  []string
	}{
		{"kms-server", []string{
			"sts.googleapis.com", "cloudkms.googleapis.com", "kms-emulator",
		}},
		{"mailpit-server", []string{"mailpit"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			certificate := leafCertificate(t, set.files[test.name+".crt"])
			if !slices.Equal(certificate.DNSNames, test.dns) || len(certificate.URIs) != 0 {
				t.Fatalf("certificate SANs differ: DNS=%v URI=%v", certificate.DNSNames, certificate.URIs)
			}
			if _, err := certificate.Verify(x509.VerifyOptions{
				Roots: roots, DNSName: test.dns[0], CurrentTime: now,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			}); err != nil {
				t.Fatalf("verify certificate: %v", err)
			}
			if _, err := tls.X509KeyPair(
				set.files[test.name+".crt"], set.files[test.name+".key"],
			); err != nil {
				t.Fatalf("certificate key mismatch: %v", err)
			}
			if certificate.NotAfter.Sub(now) < 89*24*time.Hour {
				t.Fatal("leaf certificate lifetime is unexpectedly short")
			}
		})
	}
}

func leafCertificate(t *testing.T, content []byte) *x509.Certificate {
	t.Helper()
	block, _ := pem.Decode(content)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatal("certificate PEM is invalid")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return certificate
}

func certificatePool(t *testing.T, content []byte) *x509.CertPool {
	t.Helper()
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(content) {
		t.Fatal("CA PEM is invalid")
	}
	return pool
}
