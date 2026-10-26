package actorassertionv1

import "errors"

var (
	ErrContract  = errors.New("IAM V1 actor assertion contract is invalid")
	ErrSignature = errors.New("IAM V1 actor assertion signature is invalid")
	ErrExpired   = errors.New("IAM V1 actor assertion is expired")
	ErrClosed    = errors.New("IAM V1 actor assertion signer is closed")
)
