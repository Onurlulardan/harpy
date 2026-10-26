package sessionview

import "errors"

var (
	ErrContract = errors.New("IAM session view contract is invalid")
	ErrRejected = errors.New("IAM session view is rejected")
	ErrStorage  = errors.New("IAM session view storage is unavailable")
)
