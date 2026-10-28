package main

import (
	"crypto/x509"
	"time"
)

func addRuntimePKI(set *artifactSet, now time.Time) error {
	authorityName := "ForgeCRUD local runtime CA"
	if set.options.profile == profileForgeGuard {
		authorityName = "ForgeCRUD host runtime CA"
	}
	authority, err := newCertificateAuthority(
		now, authorityName, 365*24*time.Hour,
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
	kmsDNSNames := []string{
		"sts.googleapis.com", "cloudkms.googleapis.com", "kms-emulator",
	}
	if set.options.profile == profileForgeGuard {
		kmsDNSNames = []string{
			"sts.googleapis.com", "cloudkms.googleapis.com", "kms-service",
		}
	}
	requests := []struct {
		name string
		dns  []string
	}{
		{"kms-server", kmsDNSNames},
		{"mailpit-server", []string{"mailpit"}},
	}
	if set.options.profile == profileForgeGuard {
		requests = append(requests,
			struct {
				name string
				dns  []string
			}{"s3-owner-server", []string{"s3-owner"}},
			struct {
				name string
				dns  []string
			}{"telemetry-server", []string{"telemetry"}},
		)
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
