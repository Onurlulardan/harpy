package accesstoken

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"sort"
)

const maximumPublicKeys = 64

type PublicKeySet struct {
	keys         map[string]ed25519.PublicKey
	document     []byte
	etag         string
	cacheControl string
}

func NewPublicKeySet(keys map[string]ed25519.PublicKey) (*PublicKeySet, error) {
	if len(keys) == 0 || len(keys) > maximumPublicKeys {
		return nil, ErrInvalidKey
	}
	keyIDs := make([]string, 0, len(keys))
	cloned := make(map[string]ed25519.PublicKey, len(keys))
	for keyID, key := range keys {
		if !validKeyID(keyID) || len(key) != ed25519.PublicKeySize || allZeroBytes(key) {
			return nil, ErrInvalidKey
		}
		keyIDs = append(keyIDs, keyID)
		cloned[keyID] = append(ed25519.PublicKey(nil), key...)
	}
	sort.Strings(keyIDs)
	document, err := marshalJWKS(keyIDs, cloned)
	if err != nil || len(document) == 0 {
		return nil, ErrInvalidKey
	}
	digest := sha256.Sum256(document)
	return &PublicKeySet{
		keys: cloned, document: document,
		etag:         `"sha256-` + encodeBase64URL(digest[:]) + `"`,
		cacheControl: "public, max-age=300, must-revalidate",
	}, nil
}

func (set *PublicKeySet) AccessTokenKey(
	ctx context.Context,
	keyID string,
) (ed25519.PublicKey, error) {
	if set == nil || ctx == nil || ctx.Err() != nil || !validKeyID(keyID) {
		return nil, ErrInvalidKey
	}
	key, exists := set.keys[keyID]
	if !exists {
		return nil, ErrInvalidKey
	}
	return append(ed25519.PublicKey(nil), key...), nil
}

func (set *PublicKeySet) JWKS() ([]byte, string, string, error) {
	if set == nil || len(set.document) == 0 || set.etag == "" {
		return nil, "", "", ErrInvalidKey
	}
	return append([]byte(nil), set.document...), set.etag, set.cacheControl, nil
}

func marshalJWKS(keyIDs []string, keys map[string]ed25519.PublicKey) ([]byte, error) {
	document := jwksDocument{Keys: make([]jwkDocument, 0, len(keyIDs))}
	for _, keyID := range keyIDs {
		document.Keys = append(document.Keys, jwkDocument{
			KeyType: "OKP", Curve: "Ed25519", X: encodeBase64URL(keys[keyID]),
			Use: "sig", KeyID: keyID, Algorithm: Algorithm,
		})
	}
	return json.Marshal(document)
}

func allZeroBytes(value []byte) bool {
	var combined byte
	for _, item := range value {
		combined |= item
	}
	return combined == 0
}
