package iam

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func mapHTTPInitialPasswordError(err error) error {
	if errors.Is(err, authentication.ErrInvalidCredentials) ||
		errors.Is(err, authentication.ErrInitialPasswordState) ||
		errors.Is(err, platformtx.ErrCredentialTarget) {
		return ErrAuthHTTPInvalidCredentials
	}
	if errors.Is(err, authentication.ErrInitialPasswordPolicy) {
		return ErrAuthHTTPPasswordRejected
	}
	return ErrAuthHTTPUnavailable
}
