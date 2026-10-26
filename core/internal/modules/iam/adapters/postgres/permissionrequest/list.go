package postgres

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

func (store *permissionRequestStore) List(
	ctx context.Context,
	scope tenantScope,
	query permissionrequest.ListQuery,
) (permissionrequest.Page, error) {
	if query.Validate() != nil ||
		!permissionRequestSnapshotScopeUsable(ctx, scope) {
		return permissionrequest.Page{}, permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return permissionrequest.Page{}, err
	}
	states := make([]string, 0, len(query.States))
	for _, state := range query.States {
		states = append(states, string(state))
	}
	statesJSON, err := json.Marshal(states)
	if err != nil {
		return permissionrequest.Page{}, permissionrequest.ErrInvalid
	}
	var after *uuid.UUID
	if query.After != uuid.Nil {
		after = &query.After
	}
	rows, err := bound.database.Query(
		ctx, permissionRequestListSQL,
		scope.ActorID(), string(statesJSON), after, query.Limit+1,
	)
	if err != nil {
		return permissionrequest.Page{}, mapPermissionRequestError(err)
	}
	defer rows.Close()
	page := permissionrequest.Page{Groups: make([]permissionrequest.Group, 0, query.Limit)}
	for rows.Next() {
		group, scanErr := scanPermissionRequestGroup(rows)
		if scanErr != nil {
			return permissionrequest.Page{}, scanErr
		}
		if len(page.Groups) == query.Limit {
			page.Next = page.Groups[len(page.Groups)-1].ID
			continue
		}
		page.Groups = append(page.Groups, group)
	}
	if err := rows.Err(); err != nil {
		return permissionrequest.Page{}, mapPermissionRequestError(err)
	}
	return page, nil
}
