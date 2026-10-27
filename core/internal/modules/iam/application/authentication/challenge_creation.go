package authentication

import (
	"time"

	"github.com/google/uuid"
)

type ChallengeCreation struct {
	ID, UserID, SessionID                              uuid.UUID
	ExpectedTokenVersion, ExpectedAuthorizationVersion int64
	AccessMode                                         AccessMode
	PortalBindingID                                    uuid.UUID
	Method                                             ReauthenticationMethod
	AuthenticatorID                                    uuid.UUID
	HMACKeyID                                          string
	ChallengeHMAC                                      [32]byte
	ContextSHA256                                      [32]byte
	PurposeSHA256                                      [32]byte
	ActionPolicyVersion                                int64
	OriginSHA256                                       [32]byte
	RequirePhishingResistant                           bool
}

type ChallengeCreated struct {
	Revision                 int64
	IssuedAt, ExpiresAt      time.Time
	CredentialID             []byte
	RPID                     string
	UserVerificationRequired bool
}

func (created ChallengeCreated) validate(method ReauthenticationMethod) error {
	if created.Revision != 1 || created.IssuedAt.IsZero() || created.ExpiresAt.IsZero() ||
		!created.ExpiresAt.Equal(created.IssuedAt.Add(5*time.Minute)) {
		return ErrAuthenticationStorage
	}
	if method == ReauthenticationWebAuthn {
		if len(created.CredentialID) == 0 || created.RPID == "" {
			return ErrAuthenticationStorage
		}
	} else if len(created.CredentialID) != 0 || created.RPID != "" ||
		created.UserVerificationRequired {
		return ErrAuthenticationStorage
	}
	return nil
}
