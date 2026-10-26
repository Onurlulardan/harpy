package webhookbody

import "errors"

var (
	ErrConfiguration = errors.New("billing webhook body cryptography is misconfigured")
	ErrBody          = errors.New("billing webhook body is invalid")
	ErrEnvelope      = errors.New("billing webhook body envelope is invalid")
	ErrCryptography  = errors.New("billing webhook body cryptography failed")
)
