package devjournal

import (
	"crypto/ed25519"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func journalRecoveryMaterial(
	control ed25519.PublicKey,
	marker byte,
) []recoveryv1.PublicMaterial {
	return []recoveryv1.PublicMaterial{
		{Purpose: recoveryv1.PurposeRecoveryManifest,
			PublicKey: journalKeyArray(journalPublic(marker))},
		{Purpose: recoveryv1.PurposeRecoveryControl, PublicKey: journalKeyArray(control)},
		{Purpose: recoveryv1.PurposeRestoreReceipt,
			PublicKey: journalKeyArray(journalPublic(marker + 1))},
	}
}

func journalKeyArray(value ed25519.PublicKey) [32]byte {
	var result [32]byte
	copy(result[:], value)
	return result
}

func journalPublic(marker byte) ed25519.PublicKey {
	public, _ := journalKey(marker)
	return public
}

func stringPointer(value string) *string { return &value }

func journalKey(marker byte) (ed25519.PublicKey, ed25519.PrivateKey) {
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = marker
	}
	private := ed25519.NewKeyFromSeed(seed)
	return private.Public().(ed25519.PublicKey), private
}
