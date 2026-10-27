package profile

import "errors"

var (
	ErrInvalidCommand   = errors.New("IAM profile schema command is invalid")
	ErrAuthorityDenied  = errors.New("IAM profile schema authority is denied")
	ErrEvidenceMissing  = errors.New("IAM profile schema evidence is missing")
	ErrReferencePresent = errors.New("IAM profile field still has live references")
	ErrConflict         = errors.New("IAM profile mutation conflicts")
	ErrSelfNotFound     = errors.New("IAM self profile was not found")
)
