package menu

import (
	"context"
	_ "embed"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

//go:embed sql/tenant/deactivate_active_configs.sql
var deactivateActiveConfigsSQL string

//go:embed sql/tenant/insert_config.sql
var insertConfigSQL string

func (storage *store) CreateConfig(
	ctx context.Context, mutation menuapp.ConfigMutation,
) (menuapp.ConfigDocument, error) {
	if !validConfigWrite(mutation) || !mutation.Active {
		return menuapp.ConfigDocument{}, menuapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return menuapp.ConfigDocument{}, unavailable()
	}
	if _, err := storage.execute(ctx, deactivateActiveConfigsSQL, mutation.ActorID.String(),
		mutation.OccurredAt); err != nil {
		return menuapp.ConfigDocument{}, unavailable()
	}
	return storage.scanConfig(storage.queryRow(ctx, insertConfigSQL, mutation.ID.String(),
		mutation.Name, []byte(mutation.Config), mutation.ActorID.String(), mutation.OccurredAt))
}
