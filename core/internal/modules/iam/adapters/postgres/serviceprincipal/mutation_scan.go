package serviceprincipal

import serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"

func scanMutation(row commandRow) (serviceapp.MutationResult, error) {
	var result serviceapp.MutationResult
	if row == nil || row.Scan(
		&result.ID, &result.Revision, &result.AuthorizationClock, &result.Replayed,
	) != nil || result.Validate() != nil {
		return serviceapp.MutationResult{}, serviceapp.ErrEvidenceMissing
	}
	return result, nil
}
