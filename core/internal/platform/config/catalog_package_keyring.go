package config

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"sort"
)

const catalogPackageKeyringEnvironmentKey = "HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING"

var catalogKeyIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,127}$`)

type CatalogPackageKeyring struct {
	keys map[string]ed25519.PublicKey
}

type catalogPackageKeyringDocument struct {
	Version int                       `json:"version"`
	Keys    []catalogPackagePublicKey `json:"keys"`
}

type catalogPackagePublicKey struct {
	KeyID     string `json:"key_id"`
	PublicKey string `json:"public_key"`
}

func loadCatalogPackageKeyring(values environmentValues) (CatalogPackageKeyring, error) {
	secret, err := loadSecret(values, catalogPackageKeyringEnvironmentKey, true, secretFileLimit)
	if err != nil {
		return CatalogPackageKeyring{}, err
	}
	document := []byte(secret.Value())
	if rejectDuplicateJSONMembers(document) != nil {
		return CatalogPackageKeyring{}, configError(catalogPackageKeyringEnvironmentKey,
			"contains ambiguous JSON")
	}
	decoder := json.NewDecoder(newStringReader(secret.Value()))
	decoder.DisallowUnknownFields()
	var raw catalogPackageKeyringDocument
	if decoder.Decode(&raw) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		raw.Version != 1 || len(raw.Keys) < 1 || len(raw.Keys) > 16 {
		return CatalogPackageKeyring{}, invalidCatalogKeyring()
	}
	keys := make(map[string]ed25519.PublicKey, len(raw.Keys))
	for _, entry := range raw.Keys {
		decoded, decodeErr := base64.RawURLEncoding.DecodeString(entry.PublicKey)
		if decodeErr != nil || !catalogKeyIDPattern.MatchString(entry.KeyID) ||
			len(decoded) != ed25519.PublicKeySize ||
			base64.RawURLEncoding.EncodeToString(decoded) != entry.PublicKey {
			return CatalogPackageKeyring{}, invalidCatalogKeyring()
		}
		if _, duplicate := keys[entry.KeyID]; duplicate {
			return CatalogPackageKeyring{}, invalidCatalogKeyring()
		}
		keys[entry.KeyID] = append(ed25519.PublicKey(nil), decoded...)
	}
	return CatalogPackageKeyring{keys: keys}, nil
}

func (keyring CatalogPackageKeyring) Resolve(keyID string) (ed25519.PublicKey, error) {
	key, found := keyring.keys[keyID]
	if !found || !catalogKeyIDPattern.MatchString(keyID) || len(key) != ed25519.PublicKeySize {
		return nil, errors.New("catalog package signing key is unavailable")
	}
	return append(ed25519.PublicKey(nil), key...), nil
}

func (keyring CatalogPackageKeyring) Validate() error {
	if len(keyring.keys) < 1 || len(keyring.keys) > 16 {
		return invalidCatalogKeyring()
	}
	for id, key := range keyring.keys {
		if !catalogKeyIDPattern.MatchString(id) || len(key) != ed25519.PublicKeySize {
			return invalidCatalogKeyring()
		}
	}
	return nil
}

// PublicKeys returns defensive copies in stable key-ID order.
func (keyring CatalogPackageKeyring) PublicKeys() []ed25519.PublicKey {
	ids := make([]string, 0, len(keyring.keys))
	for id := range keyring.keys {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	keys := make([]ed25519.PublicKey, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, append(ed25519.PublicKey(nil), keyring.keys[id]...))
	}
	return keys
}

func invalidCatalogKeyring() error {
	return configError(catalogPackageKeyringEnvironmentKey,
		"must satisfy catalog-package-public-keyring.v1")
}
