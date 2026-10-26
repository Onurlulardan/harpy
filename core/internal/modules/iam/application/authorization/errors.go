package authorization

import "errors"

var (
	ErrAuthorizationContract = errors.New("IAM authorization contract is invalid")
	ErrAuthorizationEvidence = errors.New("IAM authorization evidence is unavailable")
)
