package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
)

const maximumLocalReleaseArtifactBytes = postgresv1.MaximumMigrationManifestBytes

type artifact struct {
	name     string
	digest   string
	document []byte
}

func newArtifact(name string, value any) (artifact, error) {
	if validatable, ok := value.(interface{ Validate() error }); ok {
		if err := validatable.Validate(); err != nil {
			return artifact{}, fmt.Errorf("validate %s: %w", name, err)
		}
	}
	document, err := attestation.MarshalCanonical(value)
	if err != nil {
		return artifact{}, fmt.Errorf("encode %s: %w", name, err)
	}
	digest, err := attestation.DigestHex(document)
	if err != nil {
		return artifact{}, fmt.Errorf("digest %s: %w", name, err)
	}
	return artifact{name: name, digest: digest, document: document}, nil
}

func publishArtifact(directory string, value artifact) error {
	path := filepath.Join(directory, value.name)
	return attestation.WriteNewCanonicalBounded(
		path, value.document, maximumLocalReleaseArtifactBytes,
	)
}

type signingAuthority struct {
	signer    attestation.Signer
	trustRoot artifact
}

func newSigningAuthority(now time.Time) (signingAuthority, error) {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		return signingAuthority{}, fmt.Errorf("generate signing seed: %w", err)
	}
	defer clear(seed)
	privateKey := ed25519.NewKeyFromSeed(seed)
	keyID := "local-" + uuid.NewString()
	root, err := newArtifact("trust-root.json", map[string]any{
		"contract_version": "attestation-trust-root.v1",
		"keys": []any{map[string]any{
			"key_id": keyID, "algorithm": "Ed25519",
			"public_key": base64.RawURLEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey)),
			"not_before": now.Add(-5 * time.Minute).Format(time.RFC3339),
			"not_after":  now.Add(24 * time.Hour).Format(time.RFC3339),
		}},
	})
	if err != nil {
		return signingAuthority{}, err
	}
	parsedRoot, err := attestation.ParseTrustRoot(root.document)
	if err != nil {
		return signingAuthority{}, fmt.Errorf("parse generated trust root: %w", err)
	}
	key, err := attestation.MarshalCanonical(map[string]any{
		"contract_version": "attestation-signing-key.v1", "key_id": keyID,
		"algorithm": "Ed25519", "seed": base64.RawURLEncoding.EncodeToString(seed),
	})
	if err != nil {
		return signingAuthority{}, err
	}
	signer, err := attestation.ParseSigner(key, parsedRoot)
	clear(key)
	if err != nil {
		return signingAuthority{}, fmt.Errorf("parse generated signer: %w", err)
	}
	return signingAuthority{signer: signer, trustRoot: root}, nil
}
