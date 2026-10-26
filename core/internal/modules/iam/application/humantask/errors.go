package humantask

import "errors"

var (
	ErrInvalid     = errors.New("IAM human task access request is invalid")
	ErrUnavailable = errors.New("IAM human task access evidence is unavailable")
	ErrForbidden   = errors.New("IAM human task access is forbidden")
)
