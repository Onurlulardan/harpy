package actorassertionv1

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/url"
	"testing"
	"time"
)

type tlsFixture struct {
	now                            time.Time
	roots                          *x509.CertPool
	server, client                 tls.Certificate
	wrongServerURI, wrongClientURI tls.Certificate
}

func newTLSFixture(t *testing.T) tlsFixture {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	rootPublic, rootPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rootTemplate := &x509.Certificate{SerialNumber: big.NewInt(1),
		Subject:   pkix.Name{CommonName: "Harpy test actor bridge root"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate,
		rootPublic, rootPrivate)
	if err != nil {
		t.Fatal(err)
	}
	root, err := x509.ParseCertificate(rootDER)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(root)
	return tlsFixture{now: now, roots: roots,
		server: issueLeaf(t, root, rootPrivate, now, 2, "v1.forms.internal",
			V1EdgePeerURI(DomainDocuments), x509.ExtKeyUsageServerAuth),
		client: issueLeaf(t, root, rootPrivate, now, 3, "core-api.internal",
			CoreAPIPeerURI, x509.ExtKeyUsageClientAuth),
		wrongServerURI: issueLeaf(t, root, rootPrivate, now, 4, "v1.forms.internal",
			V1EdgePeerURI(DomainBilling), x509.ExtKeyUsageServerAuth),
		wrongClientURI: issueLeaf(t, root, rootPrivate, now, 5, "core-api.internal",
			"spiffe://harpy.internal/not-core-api", x509.ExtKeyUsageClientAuth)}
}

func issueLeaf(t *testing.T, root *x509.Certificate, rootKey ed25519.PrivateKey,
	now time.Time, serial int64, dnsName, identity string, usage x509.ExtKeyUsage) tls.Certificate {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
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
	der, err := x509.CreateCertificate(rand.Reader, template, root, public, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: private, Leaf: leaf}
}
