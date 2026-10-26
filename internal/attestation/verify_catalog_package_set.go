package attestation

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

type CatalogPackageKeyring interface {
	Resolve(string) (ed25519.PublicKey, error)
	PublicKeys() []ed25519.PublicKey
}

func VerifyCatalogPackageSet(
	document []byte,
	release VerifiedRelease,
	keyring CatalogPackageKeyring,
	releaseAuthorityKeys []ed25519.PublicKey,
) (VerifiedCatalogPackageSet, error) {
	manifest := release.Value()
	if manifest.ReleaseID == "" || keyring == nil ||
		len(releaseAuthorityKeys) == 0 {
		return VerifiedCatalogPackageSet{}, invalidCatalogPackageSet()
	}
	set, err := catalogpackagev1.DecodeCanonical(document)
	if err != nil || set.ReleaseID != manifest.ReleaseID {
		return VerifiedCatalogPackageSet{}, invalidCatalogPackageSet()
	}
	digest := sha256.Sum256(document)
	if hex.EncodeToString(digest[:]) != manifest.CatalogPackageSetSHA256 {
		return VerifiedCatalogPackageSet{}, invalidCatalogPackageSet()
	}
	packageKeys, err := catalogPackageAuthorityKeys(
		keyring.PublicKeys(), releaseAuthorityKeys,
	)
	if err != nil {
		return VerifiedCatalogPackageSet{}, err
	}
	for _, sealed := range set.Packages {
		key, resolveErr := keyring.Resolve(sealed.KeyID)
		signature, signatureErr := base64.RawURLEncoding.DecodeString(sealed.Signature)
		if resolveErr != nil || signatureErr != nil ||
			!containsCatalogKey(packageKeys, key) ||
			containsCatalogKey(releaseAuthorityKeys, key) ||
			!ed25519.Verify(key, sealed.Manifest, signature) {
			return VerifiedCatalogPackageSet{}, invalidCatalogPackageSet()
		}
	}
	return newVerifiedCatalogPackageSet(set), nil
}

func catalogPackageAuthorityKeys(
	packageKeys, releaseKeys []ed25519.PublicKey,
) ([]ed25519.PublicKey, error) {
	if len(packageKeys) < 1 || len(packageKeys) > 16 {
		return nil, invalidCatalogPackageSet()
	}
	result := make([]ed25519.PublicKey, 0, len(packageKeys))
	for _, key := range packageKeys {
		if len(key) != ed25519.PublicKeySize ||
			containsCatalogKey(result, key) ||
			containsCatalogKey(releaseKeys, key) {
			return nil, invalidCatalogPackageSet()
		}
		result = append(result, append(ed25519.PublicKey(nil), key...))
	}
	for _, key := range releaseKeys {
		if len(key) != ed25519.PublicKeySize {
			return nil, invalidCatalogPackageSet()
		}
	}
	return result, nil
}

func containsCatalogKey(keys []ed25519.PublicKey, candidate ed25519.PublicKey) bool {
	if len(candidate) != ed25519.PublicKeySize {
		return false
	}
	for _, key := range keys {
		if bytes.Equal(key, candidate) {
			return true
		}
	}
	return false
}

func invalidCatalogPackageSet() error {
	return fmt.Errorf("catalog package set verification failed")
}
