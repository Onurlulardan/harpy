package postgres

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func scanMutation(row lifecycleRow) (lifecycle.MutationResult, error) {
	var result lifecycle.MutationResult
	err := row.Scan(&result.ID, &result.SubjectUserID, &result.Revision,
		&result.TokenVersion, &result.AuthorizationVersion, &result.Replayed)
	if err != nil {
		return lifecycle.MutationResult{}, mapError(err)
	}
	if result.Validate() != nil {
		return lifecycle.MutationResult{}, lifecycle.ErrUnavailable
	}
	return result, nil
}
