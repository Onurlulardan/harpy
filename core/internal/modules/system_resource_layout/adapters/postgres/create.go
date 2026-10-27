package postgres

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

func (store *layoutStore) Create(
	ctx context.Context,
	input layoutapp.CreateStoreInput,
) (layoutapi.Revision, error) {
	if !store.commandUsable(ctx) {
		return layoutapi.Revision{}, layoutUnavailable()
	}
	var document []byte
	err := store.queryRow(
		ctx, createSQL, input.ID, string(input.Resource), input.DefinitionJSON,
		input.DefinitionSHA256[:], input.Now,
	).Scan(&document)
	if err != nil {
		return layoutapi.Revision{}, mapLayoutError(err)
	}
	return decodeRevision(document)
}
