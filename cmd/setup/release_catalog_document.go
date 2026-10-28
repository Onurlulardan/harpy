package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type catalogKeyring struct {
	id  string
	key ed25519.PublicKey
}

func (keyring catalogKeyring) Resolve(keyID string) (ed25519.PublicKey, error) {
	if keyID != keyring.id {
		return nil, errors.New("catalog package signing key is unavailable")
	}
	return append(ed25519.PublicKey(nil), keyring.key...), nil
}

func (keyring catalogKeyring) PublicKeys() []ed25519.PublicKey {
	return []ed25519.PublicKey{append(ed25519.PublicKey(nil), keyring.key...)}
}

func buildCatalogSet(keys releaseKeys) ([]byte, error) {
	payload, err := localCatalogPayload()
	if err != nil {
		return nil, err
	}
	payloadDigest := sha256.Sum256(payload)
	manifest, err := localCatalogManifest(hex.EncodeToString(payloadDigest[:]))
	if err != nil {
		return nil, err
	}
	manifestDigest := sha256.Sum256(manifest)
	signature := ed25519.Sign(keys.catalogPrivate, manifest)
	return attestation.MarshalCanonical(catalogpackagev1.ReleaseSet{
		ContractVersion: catalogpackagev1.ContractVersion, ReleaseID: localReleaseID,
		Packages: []catalogpackagev1.SealedPackage{{
			Manifest: manifest, ManifestSHA256: hex.EncodeToString(manifestDigest[:]),
			Signature: base64.RawURLEncoding.EncodeToString(signature), KeyID: catalogKeyID,
			OwnerPayloads: []catalogpackagev1.OwnerPayload{{
				ComponentID: localComponentID, Owner: "catalog",
				Contract:      catalogpackagev1.ContractCatalog,
				ContentSHA256: hex.EncodeToString(payloadDigest[:]), Document: payload,
			}},
		}},
	})
}

func verifyCatalogSet(keys releaseKeys, document []byte, release attestation.VerifiedRelease) error {
	_, err := attestation.VerifyCatalogPackageSet(
		document, release, catalogKeyring{id: catalogKeyID, key: keys.catalogPublic},
		keys.root.ActivePublicKeys(release.IssuedAt()),
	)
	return err
}

func catalogSetDigest(document []byte) string {
	sum := sha256.Sum256(document)
	return hex.EncodeToString(sum[:])
}
