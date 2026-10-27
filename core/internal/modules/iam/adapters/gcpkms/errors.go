package gcpkms

import "errors"

var (
	ErrContract    = errors.New("IAM Google Cloud KMS contract is invalid")
	ErrUnavailable = errors.New("IAM Google Cloud KMS is unavailable")
	ErrIntegrity   = errors.New("IAM Google Cloud KMS integrity check failed")
)
