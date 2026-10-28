package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"time"
)

func addQueryAndCatalogKeyrings(set *artifactSet, _ time.Time) error {
	currentID, err := randomLowerID("local-")
	if err != nil {
		return err
	}
	previousID, err := randomLowerID("local-")
	if err != nil {
		return err
	}
	currentKey, err := randomBase64(32)
	if err != nil {
		return err
	}
	previousKey, err := randomBase64(32)
	if err != nil {
		return err
	}
	if err := addJSON(set, "rest-query-cursor-keyring.json", map[string]any{
		"version":  1,
		"current":  map[string]string{"kid": currentID, "key": currentKey},
		"previous": map[string]string{"kid": previousID, "key": previousKey},
	}); err != nil {
		return err
	}
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
	if err := set.add("catalog-package-signing-seed", []byte(
		base64.RawURLEncoding.EncodeToString(seed),
	)); err != nil {
		return err
	}
	return addJSON(set, "catalog-package-public-keyring.json", map[string]any{
		"version": 1,
		"keys": []any{map[string]string{
			"key_id":     keyID,
			"public_key": base64.RawURLEncoding.EncodeToString(publicKey),
		}},
	})
}
