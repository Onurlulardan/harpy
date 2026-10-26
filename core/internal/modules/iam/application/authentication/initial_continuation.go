package authentication

import (
	"time"

	"github.com/google/uuid"
)

const (
	InitialPasswordLifetime    = 10 * time.Minute
	InitialPasswordMaxAttempts = int16(5)
)

type InitialPasswordContinuation struct {
	TenantID, UserID, InstallationOperationID uuid.UUID
	CredentialID                              uuid.UUID
	CredentialVersion, CredentialRevision     int64
	Selector                                  [16]byte
	SecretKeyID, CSRFKeyID                    string
	SecretHMAC, CSRFHMAC                      [32]byte
	State                                     string
	PolicyAttempts                            int16
	Revision                                  int64
	IssuedAt, ExpiresAt                       time.Time
}

func (continuation InitialPasswordContinuation) valid() bool {
	return continuation.TenantID != uuid.Nil && continuation.UserID != uuid.Nil &&
		continuation.InstallationOperationID != uuid.Nil &&
		continuation.CredentialID != uuid.Nil && continuation.CredentialVersion > 0 &&
		continuation.CredentialRevision > 0 &&
		continuation.Selector != [16]byte{} && validHMACKeyID(continuation.SecretKeyID) &&
		validHMACKeyID(continuation.CSRFKeyID) && continuation.SecretHMAC != [32]byte{} &&
		continuation.CSRFHMAC != [32]byte{} && continuation.State == "active" &&
		continuation.PolicyAttempts >= 0 &&
		continuation.PolicyAttempts < InitialPasswordMaxAttempts &&
		continuation.Revision > 0 && !continuation.IssuedAt.IsZero() &&
		continuation.ExpiresAt.Equal(continuation.IssuedAt.Add(InitialPasswordLifetime))
}

func (continuation InitialPasswordContinuation) binding() initialBinding {
	return initialBinding{tenantID: continuation.TenantID, userID: continuation.UserID,
		operationID:       continuation.InstallationOperationID,
		credentialID:      continuation.CredentialID,
		credentialVersion: continuation.CredentialVersion}
}

type VerifiedInitialPassword struct {
	continuation InitialPasswordContinuation
	secretHMAC   [32]byte
	csrfHMAC     [32]byte
	valid        bool
}

func (verified *VerifiedInitialPassword) Destroy() {
	verified.destroy()
}
