package actorbridge

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

type runtimeTLSFixture struct {
	now    time.Time
	client ClientTLSFiles
	server ServerTLSFiles
}

func newRuntimeTLSFixture(t *testing.T) runtimeTLSFixture {
	return newRuntimeTLSFixtureForServer(t, "v1-forms.fixture.invalid")
}

func newRuntimeTLSFixtureForServer(t *testing.T, serverName string) runtimeTLSFixture {
	t.Helper()
	directory := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	rootPublic, rootPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rootTemplate := &x509.Certificate{SerialNumber: big.NewInt(1),
		Subject:   pkix.Name{CommonName: "Harpy actor bridge test root"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign}
	rootDER, err := x509.CreateCertificate(
		rand.Reader, rootTemplate, rootTemplate, rootPublic, rootPrivate,
	)
	if err != nil {
		t.Fatal(err)
	}
	root, err := x509.ParseCertificate(rootDER)
	if err != nil {
		t.Fatal(err)
	}
	clientCert, clientKey := issueRuntimeLeaf(t, root, rootPrivate, now, 2,
		"core-api.fixture.invalid", actorassertionv1.CoreAPIPeerURI,
		x509.ExtKeyUsageClientAuth)
	serverCert, serverKey := issueRuntimeLeaf(t, root, rootPrivate, now, 3,
		serverName, actorassertionv1.V1EdgePeerURI(
			actorassertionv1.DomainForms,
		), x509.ExtKeyUsageServerAuth)
	rootPath := filepath.Join(directory, "root.crt")
	clientCertPath := filepath.Join(directory, "client.crt")
	clientKeyPath := filepath.Join(directory, "client.key")
	serverCertPath := filepath.Join(directory, "server.crt")
	serverKeyPath := filepath.Join(directory, "server.key")
	writeRuntimeTLSFile(t, rootPath, pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: rootDER},
	))
	writeRuntimeTLSFile(t, clientCertPath, clientCert)
	writeRuntimeTLSFile(t, clientKeyPath, clientKey)
	writeRuntimeTLSFile(t, serverCertPath, serverCert)
	writeRuntimeTLSFile(t, serverKeyPath, serverKey)
	return runtimeTLSFixture{now: now,
		client: ClientTLSFiles{CertificateFile: clientCertPath,
			PrivateKeyFile: clientKeyPath, ServerCAFile: rootPath},
		server: ServerTLSFiles{CertificateFile: serverCertPath,
			PrivateKeyFile: serverKeyPath, ClientCAFile: rootPath}}
}

func issueRuntimeLeaf(
	t *testing.T,
	root *x509.Certificate,
	rootKey ed25519.PrivateKey,
	now time.Time,
	serial int64,
	dnsName string,
	identity string,
	usage x509.ExtKeyUsage,
) ([]byte, []byte) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	identityURI, err := url.Parse(identity)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(serial),
		Subject: pkix.Name{CommonName: dnsName}, DNSNames: []string{dnsName},
		URIs: []*url.URL{identityURI}, NotBefore: now.Add(-time.Minute),
		NotAfter: now.Add(time.Minute), KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{usage}}
	der, err := x509.CreateCertificate(
		rand.Reader, template, root, publicKey, rootKey,
	)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
}

func writeRuntimeTLSFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
}
