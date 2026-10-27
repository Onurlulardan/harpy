package lifecycle

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/lifecyclecredential"
	"github.com/google/uuid"
)

type TokenIssue = lifecyclecredential.TokenIssue

type EmailVerificationStartCommand struct {
	Evidence     MutationEvidence
	IdentifierID uuid.UUID
	Token        TokenIssue
	Delivery     DeliveryArtifact
}

type PasswordResetStartCommand struct {
	Evidence         CredentialEvidence
	IdentifierKind   string
	IdentifierSHA256 [32]byte
	Token            TokenIssue
	Delivery         DeliveryArtifact
}

type PasswordResetCompleteCommand struct {
	Evidence CredentialEvidence
	Proof    CredentialProof
	Salt     [16]byte
	Tag      [32]byte
}

type InitialAdminSetupAcceptCommand struct {
	Evidence CredentialEvidence
	Proof    CredentialProof
	Salt     [16]byte
	Tag      [32]byte
}

type PasswordChangeCommand struct {
	Evidence                   MutationEvidence
	UserID                     uuid.UUID
	ExpectedCredentialRevision int64
	Salt                       [16]byte
	Tag                        [32]byte
}
