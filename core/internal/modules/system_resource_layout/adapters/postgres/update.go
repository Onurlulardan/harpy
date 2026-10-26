package postgres

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

func (store *layoutStore) Update(
	ctx context.Context,
	input layoutapp.UpdateStoreInput,
) (layoutapi.Revision, error) {
	if !store.commandUsable(ctx) {
		return layoutapi.Revision{}, layoutUnavailable()
	}
	var document []byte
	err := store.queryRow(
		ctx, updateSQL, input.ID, input.ExpectedRowRevision,
		input.DefinitionJSON, input.DefinitionSHA256[:], input.Now,
	).Scan(&document)
	if err != nil {
		return layoutapi.Revision{}, mapLayoutError(err)
	}
	return decodeRevision(document)
}
