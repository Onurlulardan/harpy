package actorassertionv1

import (
	"crypto/tls"
	"crypto/x509"
	"time"
)

func NewServerTLSConfig(
	certificate tls.Certificate,
	clientRoots *x509.CertPool,
	clock func() time.Time,
) (*tls.Config, error) {
	if len(certificate.Certificate) < 1 || certificate.PrivateKey == nil ||
		clientRoots == nil || clock == nil {
		return nil, ErrContract
	}
	return &tls.Config{
		MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13,
		Certificates: []tls.Certificate{certificate}, ClientCAs: clientRoots,
		ClientAuth: tls.RequireAndVerifyClientCert,
		VerifyConnection: func(state tls.ConnectionState) error {
			return VerifyMTLSPeer(state, CoreAPIPeerURI, x509.ExtKeyUsageClientAuth, clock())
		},
	}, nil
}
