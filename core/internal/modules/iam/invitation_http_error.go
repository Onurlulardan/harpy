package iam

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func mapInvitationApplicationError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, invitation.ErrInvalid):
		return ErrInvitationHTTPInvalid
	case errors.Is(err, invitation.ErrDenied):
		return ErrInvitationHTTPForbidden
	case errors.Is(err, invitation.ErrNotFound):
		return ErrInvitationHTTPNotFound
	case errors.Is(err, invitation.ErrConflict), errors.Is(err, invitation.ErrStale):
		return ErrInvitationHTTPConflict
	case errors.Is(err, invitation.ErrRecoveryRequired):
		return ErrInvitationHTTPRecoveryRequired
	case errors.Is(err, platformtx.ErrCommitRejected):
		return ErrInvitationHTTPConflict
	default:
		return ErrInvitationHTTPUnavailable
	}
}

func mapInvitationCredentialError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrInvitationHTTPInvalid) {
		return ErrInvitationHTTPInvalid
	}
	if errors.Is(err, invitation.ErrRecoveryRequired) ||
		errors.Is(err, ErrInvitationHTTPRecoveryRequired) {
		return ErrInvitationHTTPRecoveryRequired
	}
	if errors.Is(err, invitation.ErrUnavailable) ||
		errors.Is(err, platformtx.ErrScopeDatabase) {
		return ErrInvitationHTTPUnavailable
	}
	return ErrInvitationHTTPInvalidCredential
}
