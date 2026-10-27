package lifecycle

import (
	"time"

	"github.com/google/uuid"
)

type AuthenticatorListQuery struct {
	UserID                                             uuid.UUID
	ActorID, PermissionID                              uuid.UUID
	ExpectedTokenVersion, ExpectedAuthorizationVersion int64
	Limit                                              int
	AfterUpdatedAt                                     time.Time
	AfterID                                            uuid.UUID
}

type AuthenticatorRevisionCommand = RevisionCommand

type ChallengeEvidence struct {
	ID                  uuid.UUID
	HMACKeyID           string
	HMAC, ContextSHA256 [32]byte
}

type WebAuthnStartCommand struct {
	Evidence                           MutationEvidence
	AuthenticatorID, UserID, SessionID uuid.UUID
	DisplayLabel                       string
	Challenge                          ChallengeEvidence
	OriginPolicyVersion                int64
	RPIDSHA256, OriginSHA256           [32]byte
	UserVerificationRequired           bool
}

type WebAuthnCompleteCommand struct {
	Evidence                    MutationEvidence
	AuthenticatorID             uuid.UUID
	Challenge                   ChallengeEvidence
	CredentialID, PublicKeyCOSE []byte
	RPID                        string
	RPIDSHA256                  [32]byte
	OriginPolicyVersion         int64
	AAGUID                      uuid.UUID
	SignCount                   int64
	BackupEligible, BackupState bool
	Discoverable                bool
	UserVerificationRequired    bool
}
