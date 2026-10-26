package password

import "errors"

var (
	ErrPasswordInvalid           = errors.New("password is invalid")
	ErrPasswordRejected          = errors.New("password is rejected by policy")
	ErrPasswordPolicyUnavailable = errors.New("password policy is unavailable")
	ErrSecretDestroyed           = errors.New("password secret is destroyed")
	ErrCredentialInvalid         = errors.New("password credential is invalid")
	ErrLegacyCredentialExpired   = errors.New("legacy password credential is expired")
)
