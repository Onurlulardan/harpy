package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func addActorBridgeArtifacts(set *artifactSet, now time.Time) error {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	seed := privateKey.Seed()
	defer clear(seed)
	defer clear(privateKey)
	keyID, err := randomLowerID("local-")
	if err != nil {
		return err
	}
	if err := set.add("actor-bridge-signing-seed", []byte(
		base64.RawURLEncoding.EncodeToString(seed),
	)); err != nil {
		return err
	}
	if err := addJSON(set, "actor-bridge-public-keys.json", map[string]any{
		"version": 1,
		"keys": []any{map[string]string{
			"key_id":     keyID,
			"public_key": base64.RawURLEncoding.EncodeToString(publicKey),
		}},
	}); err != nil {
		return err
	}
	set.actorBridgeKeyID = keyID
	if err := addActorBridgeTargets(set); err != nil {
		return err
	}
	return addActorBridgeCertificates(set, now)
}

func addActorBridgeTargets(set *artifactSet) error {
	targets := make([]map[string]string, 0, len(actorBridgeDomains))
	for _, domain := range actorBridgeDomains {
		host := "v1-" + domain
		targets = append(targets, map[string]string{
			"domain":      domain,
			"base_url":    "https://" + host + ":8443",
			"server_name": host,
		})
	}
	return addJSON(set, "actor-bridge-targets.json", map[string]any{
		"version": 1, "targets": targets,
	})
}

func addActorBridgeCertificates(set *artifactSet, now time.Time) error {
	authority, err := newCertificateAuthority(
		now, "Harpy local actor bridge CA", 365*24*time.Hour,
	)
	if err != nil {
		return err
	}
	defer clear(authority.privateKey)
	key, err := authority.privateKeyPEM()
	if err != nil {
		return err
	}
	if err := set.add("actor-bridge-server-ca.crt", authority.certificatePEM()); err != nil {
		return err
	}
	if err := set.add("actor-bridge-server-ca.key", key); err != nil {
		return err
	}
	clientCert, clientKey, expiresAt, err := issueLeafCertificate(
		authority, now, leafCertificateRequest{
			commonName: "core-api", uriNames: []string{actorassertionv1.CoreAPIPeerURI},
			usage: x509.ExtKeyUsageClientAuth, validity: 90 * 24 * time.Hour,
		},
	)
	if err != nil {
		return err
	}
	if err := set.add("actor-bridge-client.crt", clientCert); err != nil {
		return err
	}
	if err := set.add("actor-bridge-client.key", clientKey); err != nil {
		return err
	}
	set.trackExpiry(expiresAt)
	for _, domainName := range actorBridgeDomains {
		domain := actorassertionv1.Domain(domainName)
		host := "v1-" + domainName
		certificate, privateKey, leafExpiry, issueErr := issueLeafCertificate(
			authority, now, leafCertificateRequest{
				commonName: host, dnsNames: []string{host},
				uriNames: []string{actorassertionv1.V1EdgePeerURI(domain)},
				usage:    x509.ExtKeyUsageServerAuth, validity: 90 * 24 * time.Hour,
			},
		)
		if issueErr != nil {
			return issueErr
		}
		name := "actor-bridge-server-" + domainName
		if err := set.add(name+".crt", certificate); err != nil {
			return err
		}
		if err := set.add(name+".key", privateKey); err != nil {
			return err
		}
		set.trackExpiry(leafExpiry)
	}
	set.trackExpiry(authority.certificate.NotAfter)
	return nil
}
