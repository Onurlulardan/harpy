package actorassertionv1

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"time"
)

const CoreAPIPeerURI = "spiffe://harpy.internal/core-api"

func V1EdgePeerURI(domain Domain) string {
	if domain.Validate() != nil {
		return ""
	}
	return "spiffe://harpy.internal/v1-edge/" + string(domain)
}

func VerifyMTLSPeer(
	state tls.ConnectionState,
	expectedURI string,
	usage x509.ExtKeyUsage,
	now time.Time,
) error {
	parsed, err := url.Parse(expectedURI)
	if err != nil || parsed.Scheme != "spiffe" || parsed.Host != "harpy.internal" ||
		parsed.RawQuery != "" || parsed.Fragment != "" || now.IsZero() ||
		state.Version != tls.VersionTLS13 ||
		len(state.VerifiedChains) < 1 || len(state.PeerCertificates) < 1 ||
		len(state.PeerCertificates) > 4 {
		return ErrContract
	}
	leaf := state.PeerCertificates[0]
	if now.Before(leaf.NotBefore) || !now.Before(leaf.NotAfter) || len(leaf.URIs) != 1 ||
		leaf.URIs[0].String() != expectedURI || !hasExtendedUsage(leaf, usage) {
		return ErrContract
	}
	if len(state.VerifiedChains[0]) < 2 ||
		!bytes.Equal(state.VerifiedChains[0][0].Raw, leaf.Raw) {
		return ErrContract
	}
	return nil
}

func hasExtendedUsage(certificate *x509.Certificate, expected x509.ExtKeyUsage) bool {
	for _, usage := range certificate.ExtKeyUsage {
		if usage == expected {
			return true
		}
	}
	return false
}
