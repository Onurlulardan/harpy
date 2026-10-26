package authentication

import (
	"crypto/subtle"
	"time"

	"github.com/google/uuid"
)

type ReauthenticationCredential struct {
	ChallengeID, UserID, SessionID, AuthenticatorID       uuid.UUID
	ChallengeRevision                                     int64
	AttemptCount                                          int16
	ChallengeState                                        string
	Method                                                ReauthenticationMethod
	HMACKeyID                                             string
	ChallengeHMAC                                         [32]byte
	ContextSHA256, PurposeSHA256, OriginSHA256            [32]byte
	ActionPolicyVersion                                   int64
	ExpiresAt                                             time.Time
	UserRevision, TokenVersion, AuthorizationVersion      int64
	PlatformAuthorizationClock                            int64
	SecurityEpoch                                         [32]byte
	UserStatus, IdentityKind                              string
	SessionRevision                                       int64
	SessionState                                          string
	SessionAbsoluteExpiresAt                              time.Time
	AccessModeID                                          uuid.UUID
	AccessMode                                            AccessMode
	AccessModeState                                       string
	AccessModeRevision                                    int64
	PortalBindingID                                       uuid.UUID
	PortalBindingState                                    string
	PortalBindingRevision                                 int64
	FamilyID, TokenID                                     uuid.UUID
	FamilyRevision, TokenRevision                         int64
	FamilyState, TokenState                               string
	FamilyAbsoluteExpiresAt                               time.Time
	PasswordCredentialVersion, PasswordCredentialRevision int64
	Password                                              PasswordCredential
	AuthenticatorState, AuthenticatorKind                 string
	AuthenticatorRevision                                 int64
	WebAuthn                                              WebAuthnCredential
	TOTP                                                  TOTPEnvelope
}

func (credential ReauthenticationCredential) validate(
	lookup ReauthenticationLookup, now time.Time,
) error {
	if credential.ChallengeID != lookup.ChallengeID ||
		credential.UserID != lookup.UserID || credential.SessionID != lookup.SessionID ||
		credential.ChallengeRevision <= 0 || credential.AttemptCount < 0 ||
		credential.AttemptCount >= 5 || credential.ChallengeState != "pending" ||
		credential.Method != lookup.Method ||
		!validHMACKeyID(credential.HMACKeyID) ||
		credential.ChallengeHMAC == [32]byte{} ||
		subtle.ConstantTimeCompare(credential.ContextSHA256[:], lookup.ContextSHA256[:]) != 1 ||
		credential.PurposeSHA256 != lookup.PurposeSHA256 ||
		credential.OriginSHA256 != lookup.OriginSHA256 ||
		credential.ActionPolicyVersion != lookup.ActionPolicyVersion ||
		credential.ExpiresAt.IsZero() || !now.Before(credential.ExpiresAt) ||
		credential.baseInvalid(lookup) || credential.proofInvalid() {
		return ErrAuthenticationStorage
	}
	return nil
}
