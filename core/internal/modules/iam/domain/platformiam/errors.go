package platformiam

import "errors"

var (
	ErrInvalidCommand = errors.New("platform IAM command is invalid")
	ErrInvalidQuery   = errors.New("platform IAM query is invalid")
	ErrDenied         = errors.New("platform IAM authority is denied")
	ErrConflict       = errors.New("platform IAM state conflicts with the request")
	ErrNotFound       = errors.New("platform IAM state was not found")
	ErrStorage        = errors.New("platform IAM storage failed")
)
