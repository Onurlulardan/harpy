package lifecycle

import "github.com/google/uuid"

type IdentifierAddCommand struct {
	Evidence                            MutationEvidence
	ID, UserID                          uuid.UUID
	Kind, DisplayValue, NormalizedValue string
	NormalizedSHA256                    [32]byte
}

type CredentialProof struct {
	Selector                  [16]byte
	SecretHMAC, BindingSHA256 [32]byte
	HashKeyID                 string
}

func (proof CredentialProof) Validate() error {
	if proof.Selector == [16]byte{} || proof.SecretHMAC == [32]byte{} ||
		proof.BindingSHA256 == [32]byte{} || proof.HashKeyID == "" {
		return ErrInvalid
	}
	return nil
}

type IdentifierVerifyCommand struct {
	Evidence CredentialEvidence
	Proof    CredentialProof
}

type RevisionCommand struct {
	Evidence         MutationEvidence
	ID               uuid.UUID
	ExpectedRevision int64
}
