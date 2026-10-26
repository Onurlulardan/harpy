//go:build integration

package postgres

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type postgresTLSAuthority struct {
	caPath     string
	serverPath string
	keyPath    string
}

func writePostgresTLSAuthority(
	t *testing.T,
	directory string,
	prefix string,
) postgresTLSAuthority {
	t.Helper()
	caPrivate, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	notBefore := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	notAfter := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	ca := x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: prefix + " root"},
		NotBefore: notBefore, NotAfter: notAfter, IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	caDER, err := x509.CreateCertificate(
		rand.Reader, &ca, &ca, &caPrivate.PublicKey, caPrivate,
	)
	if err != nil {
		t.Fatal(err)
	}
	serverPrivate, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "localhost"},
		NotBefore: notBefore, NotAfter: notAfter, DNSNames: []string{"localhost"},
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverDER, err := x509.CreateCertificate(
		rand.Reader, &server, &ca, &serverPrivate.PublicKey, caPrivate,
	)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(serverPrivate)
	if err != nil {
		t.Fatal(err)
	}
	authority := postgresTLSAuthority{
		caPath:     filepath.Join(directory, prefix+"-ca.pem"),
		serverPath: filepath.Join(directory, prefix+"-server.pem"),
		keyPath:    filepath.Join(directory, prefix+"-server-key.pem"),
	}
	writePostgresTLSPEM(t, authority.caPath, "CERTIFICATE", caDER)
	writePostgresTLSPEM(t, authority.serverPath, "CERTIFICATE", serverDER)
	writePostgresTLSPEM(t, authority.keyPath, "PRIVATE KEY", keyDER)
	return authority
}

func writePostgresTLSPEM(t *testing.T, path string, kind string, der []byte) {
	t.Helper()
	document := pem.EncodeToMemory(&pem.Block{Type: kind, Bytes: der})
	if document == nil {
		t.Fatal("encode PostgreSQL TLS fixture")
	}
	if err := os.WriteFile(path, document, 0o600); err != nil {
		t.Fatal(err)
	}
}
