package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

const (
	releaseKeyID         = "local-release"
	catalogKeyID         = "local-catalog"
	localDatabase        = "harpy"
	localSystemID        = "123456789"
	localDatabaseOID     = 16384
	localReleaseID       = "11111111-1111-4111-8111-111111111111"
	localEnvironmentID   = "22222222-2222-4222-8222-222222222222"
	localAuthorityID     = "33333333-3333-4333-8333-333333333333"
	localOwnershipID     = "33333333-3333-4333-8333-333333333334"
	localClusterLease    = "55555555-5555-4555-8555-555555555555"
	localExtensionLease  = "66666666-6666-4666-8666-666666666666"
	localBootstrapRun    = "77777777-7777-4777-8777-777777777777"
	localExtensionRun    = "88888888-8888-4888-8888-888888888888"
	localCleanupRun      = "99999999-9999-4999-8999-999999999999"
	localOperationID     = "10000000-0000-4000-8000-000000000021"
	localReceiptID       = "10000000-0000-4000-8000-000000000024"
	localPackageID       = "10000000-0000-4000-8000-000000000031"
	localComponentID     = "10000000-0000-4000-8000-000000000032"
	localFixedDigest     = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	localReleaseValidity = 24 * time.Hour
)

type releaseClock struct {
	now, issued, expires, releaseExpires, adopted, cleaned time.Time
}

func newReleaseClock(now time.Time) releaseClock {
	now = now.UTC().Truncate(time.Second)
	issued := now.Add(-2 * time.Minute)
	return releaseClock{
		now: now, issued: issued, expires: issued.Add(12 * time.Minute),
		releaseExpires: now.Add(localReleaseValidity),
		adopted:        issued.Add(time.Minute), cleaned: issued.Add(-time.Second),
	}
}

func (clock releaseClock) stamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

type releaseKeys struct {
	clock          releaseClock
	private        ed25519.PrivateKey
	root           attestation.TrustRoot
	rootDocument   []byte
	catalogPrivate ed25519.PrivateKey
	catalogPublic  ed25519.PublicKey
	keyring        []byte
}

type trustedReleaseKey struct {
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
	NotBefore string `json:"not_before"`
	NotAfter  string `json:"not_after"`
}

type trustedReleaseRoot struct {
	ContractVersion string              `json:"contract_version"`
	Keys            []trustedReleaseKey `json:"keys"`
}

type detachedSignature struct {
	ContractVersion string                         `json:"contract_version"`
	ManifestSHA256  string                         `json:"manifest_sha256"`
	Signature       runtimev1.AttestationSignature `json:"signature"`
}

type releaseDocument struct {
	body   []byte
	digest string
}

func newReleaseKeys(clock releaseClock) (releaseKeys, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return releaseKeys{}, err
	}
	catalogPublic, catalogPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return releaseKeys{}, err
	}
	if public.Equal(catalogPublic) {
		return releaseKeys{}, fmt.Errorf("catalog signing key matches the release trust root")
	}
	rootDocument, err := attestation.MarshalCanonical(trustedReleaseRoot{
		ContractVersion: "attestation-trust-root.v1",
		Keys: []trustedReleaseKey{{
			KeyID: releaseKeyID, Algorithm: "Ed25519",
			PublicKey: base64.RawURLEncoding.EncodeToString(public),
			NotBefore: clock.stamp(clock.issued.Add(-time.Hour)),
			NotAfter:  clock.stamp(clock.releaseExpires.Add(time.Hour)),
		}},
	})
	if err != nil {
		return releaseKeys{}, err
	}
	root, err := attestation.ParseTrustRoot(rootDocument)
	if err != nil {
		return releaseKeys{}, err
	}
	keyring, err := json.Marshal(struct {
		Version int `json:"version"`
		Keys    []struct {
			KeyID     string `json:"key_id"`
			PublicKey string `json:"public_key"`
		} `json:"keys"`
	}{
		Version: 1,
		Keys: []struct {
			KeyID     string `json:"key_id"`
			PublicKey string `json:"public_key"`
		}{{
			KeyID:     catalogKeyID,
			PublicKey: base64.RawURLEncoding.EncodeToString(catalogPublic),
		}},
	})
	if err != nil {
		return releaseKeys{}, err
	}
	return releaseKeys{
		clock: clock, private: private, root: root, rootDocument: rootDocument,
		catalogPrivate: catalogPrivate, catalogPublic: catalogPublic, keyring: keyring,
	}, nil
}

func canonicalDocument(value any) (releaseDocument, error) {
	body, err := attestation.MarshalCanonical(value)
	if err != nil {
		return releaseDocument{}, err
	}
	digest, err := attestation.DigestHex(body)
	if err != nil {
		return releaseDocument{}, err
	}
	return releaseDocument{body: body, digest: digest}, nil
}

func (keys releaseKeys) signDetached(contract string, document []byte) ([]byte, error) {
	digest, err := attestation.DigestHex(document)
	if err != nil {
		return nil, err
	}
	canonical, err := attestation.Canonicalize(document)
	if err != nil {
		return nil, err
	}
	value := base64.RawURLEncoding.EncodeToString(ed25519.Sign(keys.private, canonical))
	return attestation.MarshalCanonical(detachedSignature{
		ContractVersion: contract, ManifestSHA256: digest,
		Signature: runtimev1.AttestationSignature{
			Algorithm: "Ed25519", KeyID: releaseKeyID, Value: value,
		},
	})
}

func signPayload[T any](
	keys releaseKeys, contract, kind string, payload T,
) (runtimev1.AttestationSignature, error) {
	type envelope struct {
		ContractVersion string `json:"contract_version"`
		Kind            string `json:"kind"`
		Payload         T      `json:"payload"`
	}
	document, err := attestation.MarshalCanonical(envelope{
		ContractVersion: contract, Kind: kind, Payload: payload,
	})
	if err != nil {
		return runtimev1.AttestationSignature{}, err
	}
	canonical, err := attestation.Canonicalize(document)
	if err != nil {
		return runtimev1.AttestationSignature{}, err
	}
	return runtimev1.AttestationSignature{
		Algorithm: "Ed25519", KeyID: releaseKeyID,
		Value: base64.RawURLEncoding.EncodeToString(ed25519.Sign(keys.private, canonical)),
	}, nil
}
