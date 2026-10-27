package iam

import (
	"errors"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func profileApplicationError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrProfileHTTPInvalid),
		errors.Is(err, ErrProfileHTTPFieldForbidden),
		errors.Is(err, ErrProfileHTTPForbidden),
		errors.Is(err, ErrProfileHTTPNotFound),
		errors.Is(err, ErrProfileHTTPConflict),
		errors.Is(err, ErrProfileHTTPRecoveryRequired),
		errors.Is(err, ErrProfileHTTPRecoveryBudgetExceeded),
		errors.Is(err, ErrProfileHTTPUnavailable):
		return err
	case errors.Is(err, profileapp.ErrInvalidCommand),
		errors.Is(err, profileapp.ErrFieldInvalid),
		errors.Is(err, profileapp.ErrSchemaInvalid),
		errors.Is(err, profileapp.ErrValidationInvalid),
		errors.Is(err, profileapp.ErrDocumentInvalid),
		errors.Is(err, profileapp.ErrPatchInvalid),
		errors.Is(err, profileapp.ErrRequiredFieldMissing),
		errors.Is(err, profileapp.ErrFieldUnavailable):
		return ErrProfileHTTPInvalid
	case errors.Is(err, profileapp.ErrAuthorityDenied):
		return ErrProfileHTTPForbidden
	case errors.Is(err, profileapp.ErrRecoveryRequired):
		return ErrProfileHTTPRecoveryRequired
	case errors.Is(err, profilebackfillapp.ErrBudgetExceeded):
		return ErrProfileHTTPRecoveryBudgetExceeded
	case errors.Is(err, profileapp.ErrReferencePresent),
		errors.Is(err, profileapp.ErrConflict),
		errors.Is(err, profilebackfillapp.ErrConflict),
		errors.Is(err, ErrProfileBackfillUnmappable):
		return ErrProfileHTTPConflict
	case errors.Is(err, profileapp.ErrSchemaNotFound),
		errors.Is(err, profileapp.ErrDocumentNotFound),
		errors.Is(err, profileapp.ErrSelfNotFound):
		return ErrProfileHTTPNotFound
	case errors.Is(err, profilebackfillapp.ErrInvalid),
		errors.Is(err, profilebackfillapp.ErrUnavailable),
		errors.Is(err, profilebackfillapp.ErrEffects):
		return ErrProfileHTTPUnavailable
	default:
		return ErrProfileHTTPUnavailable
	}
}
