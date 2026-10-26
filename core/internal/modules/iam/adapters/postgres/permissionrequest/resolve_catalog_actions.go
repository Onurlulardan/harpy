package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (store *permissionRequestStore) ResolveCatalogActions(
	ctx context.Context,
	scope tenantScope,
	query permissionrequest.CatalogQuery,
) (permissionrequest.CatalogActionPage, error) {
	if query.Validate() != nil || query.Kind != permissionrequest.CatalogKindAll ||
		!permissionRequestSnapshotScopeUsable(ctx, scope) {
		return permissionrequest.CatalogActionPage{}, permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return permissionrequest.CatalogActionPage{}, err
	}
	rows, err := bound.database.Query(
		ctx, permissionRequestCatalogActionsSQL, scope.ActorID(), query.Search,
		nullableCatalogID(query.After), query.Limit+1,
	)
	if err != nil {
		return permissionrequest.CatalogActionPage{}, mapPermissionRequestError(err)
	}
	defer rows.Close()
	page := permissionrequest.CatalogActionPage{
		Items: make([]permissionrequest.CatalogAction, 0, query.Limit),
	}
	for rows.Next() {
		var item permissionrequest.CatalogAction
		if err := rows.Scan(
			&item.ID, &item.Key, &item.DisplayName, &item.Description,
		); err != nil {
			return permissionrequest.CatalogActionPage{}, mapPermissionRequestError(err)
		}
		if len(page.Items) == query.Limit {
			page.Next = page.Items[len(page.Items)-1].ID
			continue
		}
		page.Items = append(page.Items, item)
	}
	if err := rows.Err(); err != nil {
		return permissionrequest.CatalogActionPage{}, mapPermissionRequestError(err)
	}
	return page, nil
}
