package postgres

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
)

type postgresStateError interface{ SQLState() string }

func mapInvitationError(err error) error {
	if err == nil {
		return nil
	}
	var postgresError postgresStateError
	if !errors.As(err, &postgresError) {
		return invitation.ErrUnavailable
	}
	switch postgresError.SQLState() {
	case "P5741":
		return invitation.ErrRecoveryRequired
	case "P0002":
		return invitation.ErrNotFound
	case "40001":
		return invitation.ErrStale
	case "42501":
		return invitation.ErrDenied
	case "23505":
		return invitation.ErrConflict
	case "22023", "23514", "23503", "23502":
		return invitation.ErrInvalid
	default:
		return invitation.ErrUnavailable
	}
}
