package actorassertionv1

import (
	"crypto/tls"
	"crypto/x509"
	"time"
)

func NewClientTLSConfig(
	certificate tls.Certificate,
	roots *x509.CertPool,
	serverName string,
	target Domain,
	clock func() time.Time,
) (*tls.Config, error) {
	if len(certificate.Certificate) < 1 || certificate.PrivateKey == nil || roots == nil ||
		serverName == "" || target.Validate() != nil || clock == nil {
		return nil, ErrContract
	}
	expectedPeer := V1EdgePeerURI(target)
	return &tls.Config{
		MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13,
		RootCAs: roots, ServerName: serverName,
		Certificates: []tls.Certificate{certificate},
		VerifyConnection: func(state tls.ConnectionState) error {
			return VerifyMTLSPeer(state, expectedPeer, x509.ExtKeyUsageServerAuth, clock())
		},
	}, nil
}
