package main

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func TestGeneratedActorBridgeCertificatesCarryExactPeerIdentities(t *testing.T) {
	set, now := generatedSetForTest(t)
	roots := certificatePool(t, set.files["actor-bridge-server-ca.crt"])
	client := leafCertificate(t, set.files["actor-bridge-client.crt"])
	if len(client.URIs) != 1 || client.URIs[0].String() != actorassertionv1.CoreAPIPeerURI ||
		len(client.DNSNames) != 0 {
		t.Fatalf("client identity differs: DNS=%v URI=%v", client.DNSNames, client.URIs)
	}
	if _, err := client.Verify(x509.VerifyOptions{
		Roots: roots, CurrentTime: now,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}); err != nil {
		t.Fatalf("verify actor client: %v", err)
	}
	if _, err := tls.X509KeyPair(
		set.files["actor-bridge-client.crt"], set.files["actor-bridge-client.key"],
	); err != nil {
		t.Fatalf("actor client key mismatch: %v", err)
	}
	for _, domainName := range actorBridgeDomains {
		domain := actorassertionv1.Domain(domainName)
		name := "actor-bridge-server-" + domainName
		certificate := leafCertificate(t, set.files[name+".crt"])
		wantURI := actorassertionv1.V1EdgePeerURI(domain)
		wantDNS := "v1-" + domainName
		if len(certificate.URIs) != 1 || certificate.URIs[0].String() != wantURI ||
			len(certificate.DNSNames) != 1 || certificate.DNSNames[0] != wantDNS {
			t.Fatalf("actor server identity differs for %s", domainName)
		}
		if _, err := certificate.Verify(x509.VerifyOptions{
			Roots: roots, DNSName: wantDNS, CurrentTime: now,
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}); err != nil {
			t.Fatalf("verify actor server %s: %v", domainName, err)
		}
		if _, err := tls.X509KeyPair(set.files[name+".crt"], set.files[name+".key"]); err != nil {
			t.Fatalf("actor server key mismatch %s: %v", domainName, err)
		}
	}
}
