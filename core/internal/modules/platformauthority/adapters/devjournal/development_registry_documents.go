package devjournal

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

type developmentTrustRoot struct {
	ContractVersion string               `json:"contract_version"`
	Keys            []developmentRootKey `json:"keys"`
}

type developmentRootKey struct {
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
	NotBefore string `json:"not_before"`
	NotAfter  string `json:"not_after"`
}

func buildDevelopmentReleaseRoot(
	public ed25519.PublicKey,
	notBefore time.Time,
	notAfter time.Time,
) ([]byte, error) {
	return json.Marshal(developmentTrustRoot{
		ContractVersion: "attestation-trust-root.v1",
		Keys: []developmentRootKey{{
			KeyID: "local-authority-release-1", Algorithm: artifact.AlgorithmEd25519,
			PublicKey: base64.RawURLEncoding.EncodeToString(public),
			NotBefore: notBefore.Format(time.RFC3339),
			NotAfter:  notAfter.Format(time.RFC3339),
		}},
	})
}

func buildDevelopmentAckTrust(
	private ed25519.PrivateKey,
	public ed25519.PublicKey,
	notBefore time.Time,
	signUntil time.Time,
) ([]byte, error) {
	document, err := artifact.SignTrustBundle(artifact.TrustBundle{
		Schema: artifact.TrustBundleSchema, RegistryKind: artifact.RegistryDevelopment,
		TrustGeneration: 1, Active: artifact.TrustKey{
			KID: developmentAckKID, Algorithm: artifact.AlgorithmEd25519,
			PublicKey: base64.RawURLEncoding.EncodeToString(public),
			NotBefore: notBefore.Format(time.RFC3339),
			SignUntil: signUntil.Format(time.RFC3339),
		},
		Retired: []artifact.RetiredTrustKey{},
	}, private)
	if err != nil {
		return nil, fmt.Errorf("sign development authority trust bundle: %w", err)
	}
	return document, nil
}

func developmentReleaseRootMatches(
	path string,
	public ed25519.PublicKey,
	now time.Time,
) (bool, error) {
	document, err := readDevelopmentRegular(path, 64*1024)
	if err != nil {
		return false, fmt.Errorf("read development authority release root: %w", err)
	}
	defer clear(document)
	var root developmentTrustRoot
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&root); err != nil {
		return false, fmt.Errorf("parse development authority release root: %w", err)
	}
	var trailing any
	if err = decoder.Decode(&trailing); err != io.EOF {
		return false, fmt.Errorf("parse development authority release root")
	}
	canonical, err := json.Marshal(root)
	if err != nil || !bytes.Equal(canonical, document) ||
		root.ContractVersion != "attestation-trust-root.v1" ||
		len(root.Keys) != 1 {
		return false, fmt.Errorf("development authority release root is invalid")
	}
	key := root.Keys[0]
	notBefore, beforeErr := time.Parse(time.RFC3339Nano, key.NotBefore)
	notAfter, afterErr := time.Parse(time.RFC3339Nano, key.NotAfter)
	decoded, decodeErr := base64.RawURLEncoding.DecodeString(key.PublicKey)
	if key.KeyID != "local-authority-release-1" ||
		key.Algorithm != artifact.AlgorithmEd25519 ||
		beforeErr != nil || afterErr != nil ||
		key.NotBefore[len(key.NotBefore)-1:] != "Z" ||
		key.NotAfter[len(key.NotAfter)-1:] != "Z" ||
		now.Before(notBefore) || !now.Before(notAfter) ||
		decodeErr != nil || len(decoded) != ed25519.PublicKeySize {
		clear(decoded)
		return false, fmt.Errorf("development authority release root is invalid")
	}
	matches := ed25519.PublicKey(decoded).Equal(public)
	clear(decoded)
	return matches, nil
}
