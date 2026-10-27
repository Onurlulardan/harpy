package authentication

import (
	"time"

	"github.com/google/uuid"
)

type ReauthenticationProof struct {
	Method                 ReauthenticationMethod
	TOTPTimeStep           int64
	WebAuthnSignCount      int64
	WebAuthnBackupEligible bool
	WebAuthnBackupState    bool
}

type ReauthenticationCompletion struct {
	ChallengeID                                        uuid.UUID
	ExpectedChallengeRevision                          int64
	ExpectedAttemptCount                               int16
	UserID, SessionID                                  uuid.UUID
	ExpectedUserRevision, ExpectedSessionRevision      int64
	ExpectedTokenVersion, ExpectedAuthorizationVersion int64
	AccessModeID                                       uuid.UUID
	AccessMode                                         AccessMode
	ExpectedAccessModeRevision                         int64
	PortalBindingID                                    uuid.UUID
	ExpectedPortalBindingRevision                      int64
	FamilyID, TokenID                                  uuid.UUID
	ExpectedFamilyRevision, ExpectedTokenRevision      int64
	AuthenticatorID                                    uuid.UUID
	ExpectedAuthenticatorRevision                      int64
	ExpectedSubtypeRevision                            int64
	ExpectedPasswordCredentialVersion                  int64
	ExpectedPasswordCredentialRevision                 int64
	PurposeSHA256, ContextSHA256                       [32]byte
	Proof                                              ReauthenticationProof
	SuccessorID                                        uuid.UUID
	SuccessorSelector                                  [16]byte
	SuccessorHashKeyID                                 string
	SuccessorSecretHMAC                                [32]byte
}

type ReauthenticationCompleted struct {
	IssuedAt                                           time.Time
	ChallengeRevision, FamilyRevision, SessionRevision int64
	ConsumedTokenRevision, SuccessorTokenRevision      int64
}

func (completed ReauthenticationCompleted) validate(
	credential ReauthenticationCredential,
) error {
	if completed.IssuedAt.IsZero() ||
		completed.ChallengeRevision != credential.ChallengeRevision+1 ||
		completed.FamilyRevision != credential.FamilyRevision+1 ||
		completed.SessionRevision != credential.SessionRevision+1 ||
		completed.ConsumedTokenRevision != credential.TokenRevision+1 ||
		completed.SuccessorTokenRevision != 1 {
		return ErrAuthenticationStorage
	}
	return nil
}

func (credential ReauthenticationCredential) completion(
	proof ReauthenticationProof,
	successorID uuid.UUID,
	selector [16]byte,
	keyID string,
	secretHMAC [32]byte,
) ReauthenticationCompletion {
	return ReauthenticationCompletion{
		ChallengeID:               credential.ChallengeID,
		ExpectedChallengeRevision: credential.ChallengeRevision,
		ExpectedAttemptCount:      credential.AttemptCount,
		UserID:                    credential.UserID, SessionID: credential.SessionID,
		ExpectedUserRevision:          credential.UserRevision,
		ExpectedSessionRevision:       credential.SessionRevision,
		ExpectedTokenVersion:          credential.TokenVersion,
		ExpectedAuthorizationVersion:  credential.AuthorizationVersion,
		AccessModeID:                  credential.AccessModeID,
		AccessMode:                    credential.AccessMode,
		ExpectedAccessModeRevision:    credential.AccessModeRevision,
		PortalBindingID:               credential.PortalBindingID,
		ExpectedPortalBindingRevision: credential.PortalBindingRevision,
		FamilyID:                      credential.FamilyID, TokenID: credential.TokenID,
		ExpectedFamilyRevision:             credential.FamilyRevision,
		ExpectedTokenRevision:              credential.TokenRevision,
		AuthenticatorID:                    credential.AuthenticatorID,
		ExpectedAuthenticatorRevision:      credential.AuthenticatorRevision,
		ExpectedSubtypeRevision:            credential.subtypeRevision(),
		ExpectedPasswordCredentialVersion:  credential.PasswordCredentialVersion,
		ExpectedPasswordCredentialRevision: credential.PasswordCredentialRevision,
		PurposeSHA256:                      credential.PurposeSHA256,
		ContextSHA256:                      credential.ContextSHA256, Proof: proof,
		SuccessorID: successorID, SuccessorSelector: selector,
		SuccessorHashKeyID: keyID, SuccessorSecretHMAC: secretHMAC,
	}
}

func (credential ReauthenticationCredential) subtypeRevision() int64 {
	if credential.Method == ReauthenticationWebAuthn {
		return credential.WebAuthn.Revision
	}
	if credential.Method == ReauthenticationTOTP {
		return credential.TOTP.Revision
	}
	return 0
}
