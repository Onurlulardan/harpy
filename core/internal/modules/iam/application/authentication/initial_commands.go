package authentication

import (
	"time"

	"github.com/google/uuid"
)

type InitialPasswordClaim struct {
	InitialPasswordCredential
	Selector               [16]byte
	SecretKeyID, CSRFKeyID string
	SecretHMAC, CSRFHMAC   [32]byte
}

type InitialPasswordClaimed struct {
	Revision            int64
	IssuedAt, ExpiresAt time.Time
}

type InitialPasswordReject struct {
	Selector         [16]byte
	SecretHMAC       [32]byte
	CSRFHMAC         [32]byte
	ExpectedRevision int64
}

type InitialPasswordRejected struct {
	State          string
	PolicyAttempts int16
	Revision       int64
}

type InitialPasswordFinalize struct {
	Selector                     [16]byte
	SecretHMAC, CSRFHMAC         [32]byte
	ExpectedContinuationRevision int64
	UserID, CredentialID         uuid.UUID
	ExpectedCredentialVersion    int64
	ExpectedCredentialRevision   int64
	Salt                         [16]byte
	Tag                          [32]byte
	RequestSHA256                [32]byte
}

type InitialPasswordFinalized struct {
	UserID, AccessModeID                       uuid.UUID
	UserRevision, TokenVersion                 int64
	AuthorizationVersion, AccessModeRevision   int64
	CredentialVersion, CredentialRevision      int64
	ContinuationRevision, InstallationRevision int64
}

func (completed InitialPasswordFinalized) valid(command InitialPasswordFinalize) bool {
	return completed.UserID == command.UserID && completed.AccessModeID != uuid.Nil &&
		completed.UserRevision > 1 && completed.TokenVersion > 1 &&
		completed.AuthorizationVersion > 0 && completed.AccessModeRevision > 0 &&
		completed.CredentialVersion == command.ExpectedCredentialVersion+1 &&
		completed.CredentialRevision == command.ExpectedCredentialRevision+1 &&
		completed.ContinuationRevision == command.ExpectedContinuationRevision+1 &&
		completed.InstallationRevision > 1
}
