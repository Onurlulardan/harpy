package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
)

var localCatalogKeyIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,127}$`)

type catalogPackageAuthority struct {
	keyID      string
	privateKey ed25519.PrivateKey
	keys       map[string]ed25519.PublicKey
}

type localCatalogKeyringDocument struct {
	Version int                     `json:"version"`
	Keys    []localCatalogPublicKey `json:"keys"`
}

type localCatalogPublicKey struct {
	KeyID     string `json:"key_id"`
	PublicKey string `json:"public_key"`
}

func loadCatalogPackageAuthority(
	options prepareOptions,
) (catalogPackageAuthority, error) {
	seedText, err := readSecret(
		options.catalogSigningSeed, "Catalog package signing seed",
	)
	if err != nil {
		return catalogPackageAuthority{}, err
	}
	seed, err := base64.RawURLEncoding.DecodeString(seedText)
	if err != nil || len(seed) != ed25519.SeedSize ||
		base64.RawURLEncoding.EncodeToString(seed) != seedText {
		return catalogPackageAuthority{}, fmt.Errorf(
			"Catalog package signing seed is invalid",
		)
	}
	defer clear(seed)
	privateKey := ed25519.NewKeyFromSeed(seed)
	document, err := os.ReadFile(options.catalogKeyring)
	if err != nil {
		clear(privateKey)
		return catalogPackageAuthority{}, fmt.Errorf(
			"read Catalog package keyring: %w", err,
		)
	}
	var decoded localCatalogKeyringDocument
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&decoded) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		decoded.Version != 1 || len(decoded.Keys) < 1 || len(decoded.Keys) > 16 {
		clear(privateKey)
		return catalogPackageAuthority{}, invalidLocalCatalogAuthority()
	}
	keys := make(map[string]ed25519.PublicKey, len(decoded.Keys))
	var signerID string
	publicKey := privateKey.Public().(ed25519.PublicKey)
	for _, entry := range decoded.Keys {
		key, decodeErr := base64.RawURLEncoding.DecodeString(entry.PublicKey)
		if decodeErr != nil || !localCatalogKeyIDPattern.MatchString(entry.KeyID) ||
			len(key) != ed25519.PublicKeySize ||
			base64.RawURLEncoding.EncodeToString(key) != entry.PublicKey {
			clear(privateKey)
			return catalogPackageAuthority{}, invalidLocalCatalogAuthority()
		}
		if _, duplicate := keys[entry.KeyID]; duplicate {
			clear(privateKey)
			return catalogPackageAuthority{}, invalidLocalCatalogAuthority()
		}
		keys[entry.KeyID] = append(ed25519.PublicKey(nil), key...)
		if bytes.Equal(key, publicKey) {
			if signerID != "" {
				clear(privateKey)
				return catalogPackageAuthority{}, invalidLocalCatalogAuthority()
			}
			signerID = entry.KeyID
		}
	}
	if signerID == "" {
		clear(privateKey)
		return catalogPackageAuthority{}, invalidLocalCatalogAuthority()
	}
	return catalogPackageAuthority{
		keyID: signerID, privateKey: privateKey, keys: keys,
	}, nil
}

func (authority catalogPackageAuthority) Resolve(
	keyID string,
) (ed25519.PublicKey, error) {
	key, found := authority.keys[keyID]
	if !found || len(key) != ed25519.PublicKeySize {
		return nil, errors.New("Catalog package key is unavailable")
	}
	return append(ed25519.PublicKey(nil), key...), nil
}

func (authority catalogPackageAuthority) PublicKeys() []ed25519.PublicKey {
	ids := make([]string, 0, len(authority.keys))
	for id := range authority.keys {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]ed25519.PublicKey, 0, len(ids))
	for _, id := range ids {
		result = append(result, append(ed25519.PublicKey(nil), authority.keys[id]...))
	}
	return result
}

func (authority catalogPackageAuthority) Sign(document []byte) ([]byte, error) {
	if authority.keyID == "" ||
		len(authority.privateKey) != ed25519.PrivateKeySize {
		return nil, invalidLocalCatalogAuthority()
	}
	return ed25519.Sign(authority.privateKey, document), nil
}

func invalidLocalCatalogAuthority() error {
	return fmt.Errorf("Catalog package signing authority is invalid")
}
