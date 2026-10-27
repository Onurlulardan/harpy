package postgres

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func scanPermissionRequestReference(
	row permissionRequestRow,
) (permissionrequest.GroupRef, error) {
	var result permissionrequest.GroupRef
	var state string
	if err := row.Scan(&result.ID, &state, &result.Revision); err != nil {
		return permissionrequest.GroupRef{}, mapPermissionRequestError(err)
	}
	result.State = permissionrequest.State(state)
	if result.Validate() != nil {
		return permissionrequest.GroupRef{}, permissionrequest.ErrUnavailable
	}
	return result, nil
}
