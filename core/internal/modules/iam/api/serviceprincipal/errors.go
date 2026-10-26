package serviceprincipal

import "errors"

var (
	ErrPublicationAuthorityConflict = errors.New(
		"IAM publication authority conflicts with reserved state",
	)
	ErrPublicationAuthorityUnavailable = errors.New(
		"IAM publication authority is unavailable",
	)
)
