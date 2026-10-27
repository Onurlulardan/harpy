package menu

import (
	"context"
	_ "embed"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

const menuDefinitionID = "018f6000-0000-7000-8000-000000000001"

//go:embed sql/tenant/get_version.sql
var getVersionSQL string

//go:embed sql/tenant/list_items.sql
var listItemsSQL string

//go:embed sql/tenant/list_overrides.sql
var listOverridesSQL string

//go:embed sql/tenant/list_active_packages.sql
var listActivePackagesSQL string

func (storage *store) Get(ctx context.Context) (menuapp.Source, error) {
	if !storage.usable(ctx, false) {
		return menuapp.Source{}, unavailable()
	}
	version, err := scanVersion(storage.queryRow(ctx, getVersionSQL, menuDefinitionID))
	if err != nil {
		if storage.notFound(err) {
			return menuapp.Source{}, menuapp.ErrNotFound
		}
		return menuapp.Source{}, unavailable()
	}
	items, err := storage.listItems(ctx, version.ID.String())
	if err != nil {
		return menuapp.Source{}, err
	}
	version.Items = items
	overrides, err := storage.listOverrides(ctx, version.ID.String())
	if err != nil {
		return menuapp.Source{}, err
	}
	active, err := storage.listActivePackages(ctx)
	if err != nil {
		return menuapp.Source{}, err
	}
	return menuapp.Source{Version: version, Overrides: overrides, ActivePackages: active}, nil
}

func scanVersion(source row) (menudomain.Version, error) {
	var idRaw string
	var number, revision int64
	if err := source.Scan(&idRaw, &number, &revision); err != nil {
		return menudomain.Version{}, err
	}
	id, err := packagecontract.ParseStableID(idRaw)
	if err != nil || number < 1 || revision < 1 {
		return menudomain.Version{}, unavailable()
	}
	return menudomain.Version{ID: id, Number: uint64(number), Revision: uint64(revision)}, nil
}
