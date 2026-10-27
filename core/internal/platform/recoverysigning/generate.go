package recoverysigning

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

type generatedSet struct {
	seeds  map[recoveryv1.Purpose][]byte
	bundle recoveryv1.Bundle
}

func generate(
	ctx context.Context,
	environmentID string,
	generatedAt time.Time,
	deps dependencies,
) (generatedSet, error) {
	seeds := make(map[recoveryv1.Purpose][]byte, 3)
	public := make([]recoveryv1.PublicMaterial, 0, 3)
	purposes := []recoveryv1.Purpose{
		recoveryv1.PurposeRecoveryManifest,
		recoveryv1.PurposeRecoveryControl,
		recoveryv1.PurposeRestoreReceipt,
	}
	for _, purpose := range purposes {
		if ctx.Err() != nil {
			return generatedSet{}, ErrInterrupted
		}
		keyPublic, keyPrivate, err := ed25519.GenerateKey(deps.random)
		if err != nil {
			return generatedSet{}, fmt.Errorf("generate recovery signing key: %w", ErrInvalid)
		}
		var encoded [ed25519.PublicKeySize]byte
		copy(encoded[:], keyPublic)
		seeds[purpose] = append([]byte(nil), keyPrivate.Seed()...)
		public = append(public, recoveryv1.PublicMaterial{
			Purpose: purpose, PublicKey: encoded,
		})
	}
	bundle, err := recoveryv1.Build(environmentID, generatedAt, public)
	if err != nil {
		return generatedSet{}, ErrInvalid
	}
	return generatedSet{seeds: seeds, bundle: bundle}, nil
}
