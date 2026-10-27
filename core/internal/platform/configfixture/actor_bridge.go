package configfixture

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

const actorBridgeKeyID = "edge-2026-07"

var actorBridgeDomains = []actorassertionv1.Domain{
	actorassertionv1.DomainAnalytics, actorassertionv1.DomainBilling,
	actorassertionv1.DomainCatalog, actorassertionv1.DomainDocuments,
	actorassertionv1.DomainNotifications, actorassertionv1.DomainWork,
}

func actorBridgeFixture(directory string, now time.Time) ([]entry, error) {
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = byte(index + 71)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	paths := map[string]string{
		"seed":    filepath.Join(directory, "actor-bridge-signing-seed"),
		"public":  filepath.Join(directory, "actor-bridge-public-keys.json"),
		"targets": filepath.Join(directory, "actor-bridge-targets.json"),
		"cert":    filepath.Join(directory, "actor-bridge-client.crt"),
		"key":     filepath.Join(directory, "actor-bridge-client.key"),
		"ca":      filepath.Join(directory, "actor-bridge-server-ca.crt"),
	}
	if err := writeFixture(paths["seed"], []byte(
		base64.RawURLEncoding.EncodeToString(seed),
	)); err != nil {
		return nil, err
	}
	publicDocument := marshal(map[string]any{"version": 1, "keys": []any{
		map[string]any{"key_id": actorBridgeKeyID, "public_key": base64.RawURLEncoding.EncodeToString(privateKey[ed25519.SeedSize:])},
	}})
	if err := writeFixture(paths["public"], []byte(publicDocument)); err != nil {
		return nil, err
	}
	if err := writeActorBridgeTargets(paths["targets"]); err != nil {
		return nil, err
	}
	if err := writeActorBridgeTLS(paths, now); err != nil {
		return nil, err
	}
	clear(seed)
	clear(privateKey)
	return []entry{
		{"HARPY_IAM_V1_ACTOR_BRIDGE_SIGNING_KEY_ID", actorBridgeKeyID},
		{"HARPY_IAM_V1_ACTOR_BRIDGE_SIGNING_SEED_FILE", paths["seed"]},
		{"HARPY_IAM_V1_ACTOR_BRIDGE_PUBLIC_KEYS_FILE", paths["public"]},
		{"HARPY_IAM_V1_ACTOR_BRIDGE_TARGETS_FILE", paths["targets"]},
		{"HARPY_IAM_V1_ACTOR_BRIDGE_CLIENT_CERT_FILE", paths["cert"]},
		{"HARPY_IAM_V1_ACTOR_BRIDGE_CLIENT_KEY_FILE", paths["key"]},
		{"HARPY_IAM_V1_ACTOR_BRIDGE_SERVER_CA_FILE", paths["ca"]},
	}, nil
}

func writeActorBridgeTargets(path string) error {
	targets := make([]map[string]string, 0, len(actorBridgeDomains))
	for _, domain := range actorBridgeDomains {
		host := "v1-" + string(domain) + ".fixture.invalid"
		targets = append(targets, map[string]string{
			"domain": string(domain), "base_url": "https://" + host,
			"server_name": host,
		})
	}
	return writeFixture(path, []byte(marshal(map[string]any{
		"version": 1, "targets": targets,
	})))
}

func writeActorBridgeTLS(paths map[string]string, now time.Time) error {
	rootPublic, rootPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	canonical := now.UTC().Truncate(time.Second)
	rootTemplate := &x509.Certificate{SerialNumber: big.NewInt(1),
		Subject:   pkix.Name{CommonName: "Harpy fixture actor bridge root"},
		NotBefore: canonical.Add(-time.Hour), NotAfter: canonical.Add(24 * time.Hour),
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign}
	rootDER, err := x509.CreateCertificate(
		rand.Reader, rootTemplate, rootTemplate, rootPublic, rootPrivate,
	)
	if err != nil {
		return err
	}
	root, err := x509.ParseCertificate(rootDER)
	if err != nil {
		return err
	}
	clientPublic, clientPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	identity, _ := url.Parse(actorassertionv1.CoreAPIPeerURI)
	clientTemplate := &x509.Certificate{SerialNumber: big.NewInt(2),
		Subject: pkix.Name{CommonName: "core-api.fixture.invalid"},
		URIs:    []*url.URL{identity}, NotBefore: canonical.Add(-time.Hour),
		NotAfter: canonical.Add(24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}
	clientDER, err := x509.CreateCertificate(
		rand.Reader, clientTemplate, root, clientPublic, rootPrivate,
	)
	if err != nil {
		return err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(clientPrivate)
	if err != nil {
		return err
	}
	certificate := append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientDER}),
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootDER})...)
	if err := writeFixture(paths["cert"], certificate); err != nil {
		return err
	}
	if err := writeFixture(paths["key"], pem.EncodeToMemory(
		&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER},
	)); err != nil {
		return err
	}
	return writeFixture(paths["ca"], pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: rootDER},
	))
}

func writeFixture(path string, content []byte) error {
	return os.WriteFile(path, content, 0o600)
}
