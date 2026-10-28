package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"time"
)

type certificateAuthority struct {
	certificate *x509.Certificate
	der         []byte
	privateKey  ed25519.PrivateKey
}

type leafCertificateRequest struct {
	commonName string
	dnsNames   []string
	uriNames   []string
	usage      x509.ExtKeyUsage
	validity   time.Duration
}

func newCertificateAuthority(
	now time.Time,
	commonName string,
	validity time.Duration,
) (certificateAuthority, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return certificateAuthority{}, err
	}
	serial, err := certificateSerial()
	if err != nil {
		clear(privateKey)
		return certificateAuthority{}, err
	}
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(validity),
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLenZero:        true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		SubjectKeyId:          publicKeyID(publicKey),
	}
	der, err := x509.CreateCertificate(
		rand.Reader, template, template, publicKey, privateKey,
	)
	if err != nil {
		clear(privateKey)
		return certificateAuthority{}, err
	}
	certificate, err := x509.ParseCertificate(der)
	if err != nil {
		clear(privateKey)
		return certificateAuthority{}, err
	}
	return certificateAuthority{certificate, der, privateKey}, nil
}

func (authority certificateAuthority) certificatePEM() []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: authority.der})
}

func (authority certificateAuthority) privateKeyPEM() ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(authority.privateKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}

func issueLeafCertificate(
	authority certificateAuthority,
	now time.Time,
	request leafCertificateRequest,
) ([]byte, []byte, time.Time, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	defer clear(privateKey)
	serial, err := certificateSerial()
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	uriNames := make([]*url.URL, 0, len(request.uriNames))
	for _, value := range request.uriNames {
		parsed, parseErr := url.Parse(value)
		if parseErr != nil || parsed.String() != value {
			return nil, nil, time.Time{}, fmt.Errorf("invalid certificate URI")
		}
		uriNames = append(uriNames, parsed)
	}
	expiresAt := now.Add(request.validity)
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: request.commonName},
		DNSNames:              append([]string(nil), request.dnsNames...),
		URIs:                  uriNames,
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              expiresAt,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{request.usage},
		SubjectKeyId:          publicKeyID(publicKey),
		AuthorityKeyId:        authority.certificate.SubjectKeyId,
	}
	der, err := x509.CreateCertificate(rand.Reader, template,
		authority.certificate, publicKey, authority.privateKey)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	chain := append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		authority.certificatePEM()...)
	key := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	return chain, key, expiresAt, nil
}

func certificateSerial() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	for {
		serial, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return nil, err
		}
		if serial.Sign() > 0 {
			return serial, nil
		}
	}
}

func publicKeyID(publicKey ed25519.PublicKey) []byte {
	digest := sha256.Sum256(publicKey)
	return append([]byte(nil), digest[:20]...)
}
