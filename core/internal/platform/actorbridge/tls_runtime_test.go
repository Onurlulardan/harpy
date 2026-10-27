package actorbridge

import (
	"crypto/tls"
	"os"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func TestRuntimeTLSLoadsExactClientAndServerSPIFFEIdentities(t *testing.T) {
	fixture := newRuntimeTLSFixture(t)
	identity, err := LoadClientIdentity(fixture.client, fixture.now)
	if err != nil || identity.roots == nil || identity.certificate.Leaf == nil {
		t.Fatalf("client identity load failed: %v", err)
	}
	server, err := LoadServerTLSConfig(
		fixture.server, actorassertionv1.DomainDocuments, func() time.Time { return fixture.now },
	)
	if err != nil || server.MinVersion != tls.VersionTLS13 ||
		server.MaxVersion != tls.VersionTLS13 || server.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Fatalf("server TLS config = %#v err=%v", server, err)
	}
	if _, err := LoadServerTLSConfig(
		fixture.server, actorassertionv1.DomainWork, func() time.Time { return fixture.now },
	); !FaultIs(err, FaultUnavailable) {
		t.Fatalf("wrong server SPIFFE error = %v", err)
	}
}

func TestRuntimeTLSFailsStartupOnMalformedMaterial(t *testing.T) {
	fixture := newRuntimeTLSFixture(t)
	if err := os.WriteFile(fixture.client.PrivateKeyFile, []byte("not a key"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadClientIdentity(fixture.client, fixture.now); !FaultIs(err, FaultUnavailable) {
		t.Fatalf("malformed client key error = %v", err)
	}
	fixture = newRuntimeTLSFixture(t)
	if err := os.WriteFile(fixture.server.ClientCAFile, []byte("not a CA"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadServerTLSConfig(
		fixture.server, actorassertionv1.DomainDocuments, func() time.Time { return fixture.now },
	); !FaultIs(err, FaultUnavailable) {
		t.Fatalf("malformed client CA error = %v", err)
	}
}
