package profile

import (
	"errors"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

type postgresStateError interface{ SQLState() string }

func scanMutation(row commandRow) (profilecommand.MutationResult, error) {
	var result profilecommand.MutationResult
	if row == nil {
		return profilecommand.MutationResult{}, profilecommand.ErrEvidenceMissing
	}
	err := row.Scan(
		&result.ID,
		&result.Revision,
		&result.AuthorizationClock,
		&result.Replayed,
	)
	if err != nil {
		return profilecommand.MutationResult{}, mapCommandError(err)
	}
	if result.Validate() != nil {
		return profilecommand.MutationResult{}, profilecommand.ErrEvidenceMissing
	}
	return result, nil
}

func mapCommandError(err error) error {
	var failure postgresStateError
	if !errors.As(err, &failure) {
		return profilecommand.ErrEvidenceMissing
	}
	switch failure.SQLState() {
	case "42501":
		return profilecommand.ErrAuthorityDenied
	case "40001", "23505":
		return profilecommand.ErrConflict
	case "22023", "22P02", "23514":
		return profilecommand.ErrInvalidCommand
	case "P5404":
		return profilecommand.ErrSelfNotFound
	default:
		return profilecommand.ErrEvidenceMissing
	}
}
