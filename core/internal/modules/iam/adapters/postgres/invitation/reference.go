package postgres

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"

func scanInvitationReference(row invitationRow) (invitation.Ref, error) {
	var result invitation.Ref
	var state string
	if err := row.Scan(&result.ID, &state, &result.Revision, &result.Replayed); err != nil {
		return invitation.Ref{}, mapInvitationError(err)
	}
	result.State = invitation.State(state)
	if result.Validate() != nil {
		return invitation.Ref{}, invitation.ErrUnavailable
	}
	return result, nil
}
