package recoverykitv1

import (
	"crypto/ecdh"
	"crypto/hpke"
	"fmt"
)

const hpkeInfoDomain = "harpy.environment-recovery-kit.hpke.v1\x00"

func Seal(kit Kit, input SealInput) ([]byte, error) {
	if !validSealInput(kit, input) {
		return nil, fmt.Errorf("recovery kit seal input is invalid")
	}
	public, err := ecdh.X25519().NewPublicKey(input.ClaimPublicKey[:])
	if err != nil {
		return nil, fmt.Errorf("recovery kit claim key is invalid")
	}
	hpkePublic, err := hpke.NewDHKEMPublicKey(public)
	if err != nil {
		return nil, fmt.Errorf("recovery kit claim key is invalid")
	}
	metadata := sealMetadata(input)
	aad, err := canonicalObject(metadata)
	if err != nil {
		return nil, err
	}
	encapsulated, sender, err := hpke.NewSender(
		hpkePublic, hpke.HKDFSHA256(), hpke.AES256GCM(), hpkeInfo(aad))
	if err != nil {
		return nil, fmt.Errorf("recovery kit HPKE sender failed")
	}
	ciphertext, err := sender.Seal(aad, kit.canonical)
	if err != nil {
		return nil, fmt.Errorf("recovery kit HPKE seal failed")
	}
	wire := sealedWire{Schema: metadata.Schema, Suite: metadata.Suite,
		EnvironmentID: metadata.EnvironmentID, OperationID: metadata.OperationID,
		Generation: metadata.Generation, ClaimPublicKeySHA256: metadata.ClaimPublicKeySHA256,
		IssuedAt: metadata.IssuedAt, ExpiresAt: metadata.ExpiresAt,
		EncapsulatedKey: encode(encapsulated), Ciphertext: encode(ciphertext)}
	return canonicalObject(wire)
}

func validSealInput(kit Kit, input SealInput) bool {
	return len(kit.canonical) > 0 && kit.EnvironmentID() == input.EnvironmentID &&
		kit.OperationID() == input.OperationID && kit.Generation() == input.Generation &&
		validUUID(input.EnvironmentID) && validUUID(input.OperationID) &&
		input.Generation > 0 && input.ClaimPublicKey != [32]byte{} &&
		canonicalTime(input.IssuedAt) && canonicalTime(input.ExpiresAt) &&
		input.ExpiresAt.After(input.IssuedAt)
}

func sealMetadata(input SealInput) sealedMetadata {
	return sealedMetadata{Schema: SealedSchema, Suite: Suite,
		EnvironmentID: input.EnvironmentID, OperationID: input.OperationID,
		Generation: input.Generation, ClaimPublicKeySHA256: digestHex(input.ClaimPublicKey[:]),
		IssuedAt:  input.IssuedAt.Format(canonicalTimeFormat),
		ExpiresAt: input.ExpiresAt.Format(canonicalTimeFormat)}
}

func hpkeInfo(metadata []byte) []byte {
	return append(append([]byte(nil), []byte(hpkeInfoDomain)...), metadata...)
}
