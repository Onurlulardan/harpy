package postgres

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	"github.com/google/uuid"
)

func (store *invitationStore) List(
	ctx context.Context,
	scope tenantScope,
	query invitation.ListQuery,
) (invitation.Page, error) {
	if query.Validate() != nil || !invitationSnapshotScopeUsable(ctx, scope) {
		return invitation.Page{}, invitation.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return invitation.Page{}, err
	}
	states := make([]string, 0, len(query.States))
	for _, state := range query.States {
		states = append(states, string(state))
	}
	statesJSON, err := json.Marshal(states)
	if err != nil {
		return invitation.Page{}, invitation.ErrInvalid
	}
	var after *uuid.UUID
	if query.After != uuid.Nil {
		after = &query.After
	}
	rows, err := bound.database.Query(
		ctx, invitationListSQL,
		scope.ActorID(), string(statesJSON), after, query.Limit+1,
	)
	if err != nil {
		return invitation.Page{}, mapInvitationError(err)
	}
	defer rows.Close()
	page := invitation.Page{
		Invitations: make([]invitation.Invitation, 0, query.Limit),
	}
	for rows.Next() {
		item, scanErr := scanInvitation(rows)
		if scanErr != nil {
			return invitation.Page{}, scanErr
		}
		if len(page.Invitations) == query.Limit {
			page.Next = page.Invitations[len(page.Invitations)-1].ID
			continue
		}
		page.Invitations = append(page.Invitations, item)
	}
	if err := rows.Err(); err != nil {
		return invitation.Page{}, mapInvitationError(err)
	}
	return page, nil
}
