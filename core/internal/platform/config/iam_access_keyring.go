package config

import (
	"context"
	"crypto/ed25519"
	"errors"
	"time"
)

const iamAccessSigningInterval = 30 * 24 * time.Hour

var errInvalidIAMAccessKeyring = errors.New("IAM access-token keyring is invalid")

type IAMAccessTokenKeyring struct {
	iamKeyRedaction
	active  iamAccessSigningKey
	retired map[string]ed25519.PublicKey
	valid   bool
}

type iamAccessSigningKey struct {
	iamKeyRedaction
	kid  string
	seed [ed25519.SeedSize]byte
}

func (keyring IAMAccessTokenKeyring) Validate() error {
	if !keyring.valid || !validIAMAccessKeyID(keyring.active.kid) ||
		iamAllZero(keyring.active.seed[:]) || len(keyring.retired) > 8 {
		return errInvalidIAMAccessKeyring
	}
	for kid, key := range keyring.retired {
		if !validIAMAccessKeyID(kid) || len(key) != ed25519.PublicKeySize || iamAllZero(key) {
			return errInvalidIAMAccessKeyring
		}
	}
	return nil
}

func (keyring IAMAccessTokenKeyring) SigningKey() (
	string,
	ed25519.PrivateKey,
	error,
) {
	if err := keyring.Validate(); err != nil {
		return "", nil, err
	}
	privateKey := ed25519.NewKeyFromSeed(keyring.active.seed[:])
	return keyring.active.kid, privateKey, nil
}

func (keyring IAMAccessTokenKeyring) AccessTokenKey(
	ctx context.Context,
	kid string,
) (ed25519.PublicKey, error) {
	if ctx == nil || ctx.Err() != nil || keyring.Validate() != nil ||
		!validIAMAccessKeyID(kid) {
		return nil, errInvalidIAMAccessKeyring
	}
	if kid == keyring.active.kid {
		privateKey := ed25519.NewKeyFromSeed(keyring.active.seed[:])
		publicKey := append(ed25519.PublicKey(nil), privateKey.Public().(ed25519.PublicKey)...)
		clear(privateKey)
		return publicKey, nil
	}
	publicKey, exists := keyring.retired[kid]
	if !exists {
		return nil, errInvalidIAMAccessKeyring
	}
	return append(ed25519.PublicKey(nil), publicKey...), nil
}

func (keyring IAMAccessTokenKeyring) PublicKeys() (
	map[string]ed25519.PublicKey,
	error,
) {
	if err := keyring.Validate(); err != nil {
		return nil, err
	}
	keys := make(map[string]ed25519.PublicKey, len(keyring.retired)+1)
	for kid, key := range keyring.retired {
		keys[kid] = append(ed25519.PublicKey(nil), key...)
	}
	current, err := keyring.AccessTokenKey(context.Background(), keyring.active.kid)
	if err != nil {
		return nil, err
	}
	keys[keyring.active.kid] = current
	return keys, nil
}

func (keyring IAMAccessTokenKeyring) keyMaterials() map[string][32]byte {
	materials := make(map[string][32]byte, len(keyring.retired)+1)
	materials[keyring.active.kid] = keyring.active.seed
	for kid, publicKey := range keyring.retired {
		var material [32]byte
		copy(material[:], publicKey)
		materials[kid] = material
	}
	return materials
}

func iamAllZero(value []byte) bool {
	var combined byte
	for _, item := range value {
		combined |= item
	}
	return combined == 0
}
