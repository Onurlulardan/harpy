package main

import (
	"crypto/x509"
	"time"
)

func addRuntimePKI(set *artifactSet, now time.Time) error {
	authority, err := newCertificateAuthority(
		now, "Harpy local runtime CA", 365*24*time.Hour,
	)
	if err != nil {
		return err
	}
	defer clear(authority.privateKey)
	key, err := authority.privateKeyPEM()
	if err != nil {
		return err
	}
	if err := set.add("local-ca.crt", authority.certificatePEM()); err != nil {
		return err
	}
	if err := set.add("local-ca.key", key); err != nil {
		return err
	}
	requests := []struct {
		name string
		dns  []string
	}{
		{"kms-server", []string{
			"sts.googleapis.com", "cloudkms.googleapis.com", "kms-emulator",
		}},
		{"mailpit-server", []string{"mailpit"}},
	}
	for _, item := range requests {
		certificate, privateKey, expiresAt, issueErr := issueLeafCertificate(
			authority, now, leafCertificateRequest{
				commonName: item.dns[0], dnsNames: item.dns,
				usage: x509.ExtKeyUsageServerAuth, validity: 90 * 24 * time.Hour,
			},
		)
		if issueErr != nil {
			return issueErr
		}
		if err := set.add(item.name+".crt", certificate); err != nil {
			return err
		}
		if err := set.add(item.name+".key", privateKey); err != nil {
			return err
		}
		set.trackExpiry(expiresAt)
	}
	set.trackExpiry(authority.certificate.NotAfter)
	return nil
}
