package actorbridge

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

const actorCorePeerURI = actorassertionv1.CoreAPIPeerURI

func validLocalCertificate(
	certificate tls.Certificate,
	expectedURI string,
	usage x509.ExtKeyUsage,
	now time.Time,
) bool {
	if len(certificate.Certificate) < 1 || certificate.PrivateKey == nil ||
		certificate.Leaf == nil || now.IsZero() {
		return false
	}
	leaf := certificate.Leaf
	if now.Before(leaf.NotBefore) || !now.Before(leaf.NotAfter) ||
		len(leaf.URIs) != 1 || leaf.URIs[0].String() != expectedURI {
		return false
	}
	for _, candidate := range leaf.ExtKeyUsage {
		if candidate == usage {
			return true
		}
	}
	return false
}

func LoadServerTLSConfig(
	files ServerTLSFiles,
	domain actorassertionv1.Domain,
	now Clock,
) (*tls.Config, error) {
	if domain.Validate() != nil || now == nil {
		return nil, NewFault(FaultInvalid)
	}
	certificate, err := loadTLSCertificate(files.CertificateFile, files.PrivateKeyFile)
	timepoint := now()
	if err != nil || !validLocalCertificate(
		certificate, actorassertionv1.V1EdgePeerURI(domain),
		x509.ExtKeyUsageServerAuth, timepoint,
	) {
		return nil, NewFault(FaultUnavailable)
	}
	clientRoots, err := loadCertPool(files.ClientCAFile)
	if err != nil {
		return nil, NewFault(FaultUnavailable)
	}
	configuration, err := actorassertionv1.NewServerTLSConfig(
		certificate, clientRoots, now,
	)
	if err != nil || configuration.MinVersion != tls.VersionTLS13 ||
		configuration.MaxVersion != tls.VersionTLS13 {
		return nil, NewFault(FaultUnavailable)
	}
	return configuration, nil
}
