package menu

import (
	"context"
	_ "embed"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

//go:embed sql/tenant/deactivate_other_configs.sql
var deactivateOtherConfigsSQL string

//go:embed sql/tenant/update_config.sql
var updateConfigSQL string

func (storage *store) UpdateConfig(
	ctx context.Context, mutation menuapp.ConfigMutation,
) (menuapp.ConfigDocument, error) {
	if !validConfigWrite(mutation) {
		return menuapp.ConfigDocument{}, menuapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return menuapp.ConfigDocument{}, unavailable()
	}
	if mutation.Active {
		if _, err := storage.execute(ctx, deactivateOtherConfigsSQL, mutation.ID.String(),
			mutation.ActorID.String(), mutation.OccurredAt); err != nil {
			return menuapp.ConfigDocument{}, unavailable()
		}
	}
	return storage.scanConfig(storage.queryRow(ctx, updateConfigSQL, mutation.ID.String(),
		mutation.Name, []byte(mutation.Config), mutation.Active, mutation.ActorID.String(),
		mutation.OccurredAt))
}
