package authentication

import (
	"github.com/google/uuid"
)

type PasswordUse struct {
	UserID                       uuid.UUID
	ExpectedUserRevision         int64
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
	ExpectedCredentialVersion    int64
	ExpectedCredentialRevision   int64
	Rehash                       Argon2idCredential
}

type PasswordUsed struct {
	UserRevision         int64
	TokenVersion         int64
	AuthorizationVersion int64
	CredentialVersion    int64
	CredentialRevision   int64
}

func (used PasswordUsed) validate() error {
	if used.UserRevision <= 0 || used.TokenVersion <= 0 ||
		used.AuthorizationVersion <= 0 || used.CredentialVersion <= 0 ||
		used.CredentialRevision <= 0 {
		return ErrAuthenticationStorage
	}
	return nil
}
