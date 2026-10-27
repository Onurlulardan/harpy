package iam

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/sessionview"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func mapHTTPAuthenticationError(err error) error {
	if errors.Is(err, authentication.ErrInvalidCredentials) ||
		errors.Is(err, sessionview.ErrRejected) ||
		errors.Is(err, platformtx.ErrCredentialTarget) {
		return ErrAuthHTTPInvalidCredentials
	}
	return ErrAuthHTTPUnavailable
}

func mapHTTPSessionError(err error) error {
	if errors.Is(mapHTTPAuthenticationError(err), ErrAuthHTTPInvalidCredentials) {
		return ErrAuthHTTPInvalidSession
	}
	return ErrAuthHTTPUnavailable
}
