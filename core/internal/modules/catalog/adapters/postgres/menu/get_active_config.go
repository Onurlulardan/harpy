package menu

import (
	"context"
	_ "embed"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

//go:embed sql/tenant/get_active_config.sql
var getActiveConfigSQL string

func (storage *store) GetActiveConfig(ctx context.Context) (menuapp.ConfigDocument, error) {
	if !storage.usable(ctx, false) {
		return menuapp.ConfigDocument{}, unavailable()
	}
	return storage.scanConfig(storage.queryRow(ctx, getActiveConfigSQL))
}
