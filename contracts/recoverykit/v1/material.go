package recoverykitv1

import (
	"crypto/ed25519"
	"encoding/base64"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func (kit Kit) TrustBundle() (recoverysigningv1.Bundle, error) {
	document, err := base64.RawURLEncoding.Strict().DecodeString(
		kit.wire.RecoveryTrustBundleBase64URL,
	)
	if err != nil {
		return recoverysigningv1.Bundle{}, ErrMaterial
	}
	defer clear(document)
	bundle, err := recoverysigningv1.Parse(document)
	if err != nil || bundle.EnvironmentID() != kit.EnvironmentID() ||
		bundle.Generation() != kit.Generation() {
		return recoverysigningv1.Bundle{}, ErrMaterial
	}
	return bundle, nil
}

func (kit Kit) SigningPrivateKey(
	purpose recoverysigningv1.Purpose,
) (ed25519.PrivateKey, bool) {
	for _, key := range kit.wire.SigningKeys {
		if key.Purpose != purpose {
			continue
		}
		decoded, valid := decode(key.PrivateKey, ed25519.PrivateKeySize)
		if !valid {
			clear(decoded)
			return nil, false
		}
		return ed25519.PrivateKey(decoded), true
	}
	return nil, false
}

func (kit Kit) RecoveryPrivateKey(keyID string) ([32]byte, bool) {
	for _, key := range kit.wire.RecoveryKeys {
		if key.KeyID != keyID {
			continue
		}
		decoded, valid := decode(key.PrivateKey, 32)
		if !valid {
			clear(decoded)
			return [32]byte{}, false
		}
		var result [32]byte
		copy(result[:], decoded)
		clear(decoded)
		return result, true
	}
	return [32]byte{}, false
}

func (kit Kit) DeploymentBindingSHA256() string {
	return kit.wire.DeploymentBindingSHA256
}

func (kit *Kit) Clear() {
	if kit == nil {
		return
	}
	clear(kit.canonical)
	kit.wire = plaintextWire{}
	kit.canonical = nil
}
