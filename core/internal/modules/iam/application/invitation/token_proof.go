package invitation

import "github.com/google/uuid"

type TokenProof struct {
	ID            uuid.UUID
	Selector      [16]byte
	HashKeyID     string
	SecretHMAC    [32]byte
	BindingSHA256 [32]byte
	Revision      int64
}

func (proof TokenProof) Validate() error {
	if proof.ID == uuid.Nil || proof.Selector == [16]byte{} ||
		!validTokenKeyID(proof.HashKeyID) ||
		proof.SecretHMAC == [32]byte{} || proof.BindingSHA256 == [32]byte{} ||
		proof.Revision < 1 {
		return ErrInvalid
	}
	return nil
}
