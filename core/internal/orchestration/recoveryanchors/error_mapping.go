package recoveryanchors

import (
	"errors"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func iamSubject(value platformopsapi.RecoveryAnchorSubject) anchorapi.Subject {
	return anchorapi.Subject{EnvironmentID: value.EnvironmentID,
		TenantID: value.TenantID, ActorID: value.ActorID}
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, anchorapi.ErrInvalid):
		return platformopsapi.ErrInvalid
	case errors.Is(err, anchorapi.ErrDenied):
		return platformopsapi.ErrDenied
	case errors.Is(err, anchorapi.ErrNotFound):
		return platformopsapi.ErrNotFound
	case errors.Is(err, anchorapi.ErrConflict):
		return platformopsapi.ErrConflict
	default:
		return platformopsapi.ErrUnavailable
	}
}
