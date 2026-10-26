package invitation

import "errors"

var (
	ErrCredentialInvalid   = errors.New("IAM invitation credential is invalid")
	ErrCredentialDestroyed = errors.New("IAM invitation credential is destroyed")
	ErrCredentialRandom    = errors.New("IAM invitation credential randomness is unavailable")
)
