package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

func (store *permissionRequestStore) GetCatalogResources(
	ctx context.Context,
	scope tenantScope,
	query permissionrequest.CatalogQuery,
) (permissionrequest.CatalogResourcePage, error) {
	return store.catalogResourcePage(
		ctx, scope, query, permissionRequestCatalogResourcesSQL, nil,
	)
}

func (store *permissionRequestStore) catalogResourcePage(
	ctx context.Context,
	scope tenantScope,
	query permissionrequest.CatalogQuery,
	statement string,
	action any,
) (permissionrequest.CatalogResourcePage, error) {
	if query.Validate() != nil || !permissionRequestSnapshotScopeUsable(ctx, scope) {
		return permissionrequest.CatalogResourcePage{}, permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return permissionrequest.CatalogResourcePage{}, err
	}
	args := []any{
		scope.ActorID(), string(query.Kind), query.Search,
		nullableCatalogID(query.After), query.Limit + 1,
	}
	if action != nil {
		args = append(args, action)
	}
	rows, err := bound.database.Query(ctx, statement, args...)
	if err != nil {
		return permissionrequest.CatalogResourcePage{}, mapPermissionRequestError(err)
	}
	defer rows.Close()
	page := permissionrequest.CatalogResourcePage{
		Items: make([]permissionrequest.CatalogResource, 0, query.Limit),
	}
	for rows.Next() {
		var item permissionrequest.CatalogResource
		var kind string
		if err := rows.Scan(
			&item.ID, &item.Key, &item.DisplayName, &item.Description, &kind,
		); err != nil {
			return permissionrequest.CatalogResourcePage{}, mapPermissionRequestError(err)
		}
		item.Kind = permissionrequest.CatalogKind(kind)
		if len(page.Items) == query.Limit {
			page.Next = page.Items[len(page.Items)-1].ID
			continue
		}
		page.Items = append(page.Items, item)
	}
	if err := rows.Err(); err != nil {
		return permissionrequest.CatalogResourcePage{}, mapPermissionRequestError(err)
	}
	return page, nil
}

func nullableCatalogID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}
