package authentication

import "errors"

var (
	ErrInitialPasswordPolicy = errors.New("initial password policy rejected the candidate")
	ErrInitialPasswordState  = errors.New("initial password continuation is invalid")
)
