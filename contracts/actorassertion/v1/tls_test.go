package actorassertionv1

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMutualTLSRequiresExactCoreAndDomainPeerIdentity(t *testing.T) {
	fixture := newTLSFixture(t)
	serverTLS, err := NewServerTLSConfig(fixture.server, fixture.roots, func() time.Time {
		return fixture.now
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(writer, "ok")
	}))
	server.TLS = serverTLS
	server.StartTLS()
	defer server.Close()
	clientTLS, err := NewClientTLSConfig(fixture.client, fixture.roots,
		"v1.forms.internal", DomainDocuments, func() time.Time { return fixture.now })
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: clientTLS}}
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("mTLS response status = %d", response.StatusCode)
	}
}

func TestMutualTLSRejectsWrongCAAndWrongSAN(t *testing.T) {
	fixture := newTLSFixture(t)
	assertTLSFailure(t, fixture, fixture.wrongServerURI, fixture.client, fixture.roots)
	assertTLSFailure(t, fixture, fixture.server, fixture.wrongClientURI, fixture.roots)
	assertTLSFailure(t, fixture, fixture.server, fixture.client, x509.NewCertPool())
	assertTLSFailure(t, fixture, fixture.server, tls.Certificate{}, fixture.roots)
}

func assertTLSFailure(t *testing.T, fixture tlsFixture, serverCertificate,
	clientCertificate tls.Certificate, clientRoots *x509.CertPool) {
	t.Helper()
	serverTLS, err := NewServerTLSConfig(serverCertificate, fixture.roots,
		func() time.Time { return fixture.now })
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.TLS = serverTLS
	server.StartTLS()
	defer server.Close()
	if len(clientCertificate.Certificate) == 0 {
		clientCertificate = fixture.client
		clientTLS := &tls.Config{MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13,
			RootCAs: clientRoots, ServerName: "v1.forms.internal"}
		client := &http.Client{Transport: &http.Transport{TLSClientConfig: clientTLS}}
		if response, requestErr := client.Get(server.URL); requestErr == nil {
			_ = response.Body.Close()
			t.Fatal("mTLS accepted a client without a certificate")
		}
		return
	}
	clientTLS, err := NewClientTLSConfig(clientCertificate, clientRoots,
		"v1.forms.internal", DomainDocuments, func() time.Time { return fixture.now })
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: clientTLS}}
	if response, requestErr := client.Get(server.URL); requestErr == nil {
		_ = response.Body.Close()
		t.Fatal("mTLS accepted a wrong CA or peer identity")
	}
}
