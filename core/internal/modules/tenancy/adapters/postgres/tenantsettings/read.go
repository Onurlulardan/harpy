package tenantsettingspostgres

import (
	"context"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
)

func (storage *store) Read(ctx context.Context,
	scope settings.Scope,
) (settings.RegistrySnapshot, error) {
	if !contextUsable(ctx) || scope.TenantID() == uuidNil {
		return settings.RegistrySnapshot{}, settings.ErrInvalid
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return settings.RegistrySnapshot{}, err
	}
	var value settings.RegistrySnapshot
	err = bound.database.QueryRow(ctx, readSQL, scope.TenantID()).Scan(
		&value.ID, &value.Name, &value.Slug, &value.Status, &value.Revision,
		&value.CreatedAt, &value.UpdatedAt)
	if err != nil && bound.rowNotFound(err) {
		return settings.RegistrySnapshot{}, settings.ErrNotFound
	}
	if err != nil || value.ID != scope.TenantID() || value.Revision < 1 {
		return settings.RegistrySnapshot{}, settings.ErrUnavailable
	}
	return value, nil
}
