package maintenance

import "errors"

var (
	ErrClaimInvalid   = errors.New("IAM maintenance claim is invalid")
	ErrCommandInvalid = errors.New("IAM maintenance command is invalid")
	ErrOutcomeInvalid = errors.New("IAM maintenance outcome is invalid")
)
