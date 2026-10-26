package serviceprincipal

import "errors"

var (
	ErrInvalidPrincipal = errors.New("IAM service principal is invalid")
	ErrInvalidGrant     = errors.New("IAM capability grant is invalid")
	ErrGrantInactive    = errors.New("IAM capability grant is inactive")
	ErrGrantMismatch    = errors.New("IAM capability grant does not cover request")
	ErrEvidenceMissing  = errors.New("IAM capability grant evidence is unavailable")
	ErrChangeSink       = errors.New("IAM service principal change sink failed")
)
