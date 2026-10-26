package postgres

import (
	"context"

	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

func (store *layoutStore) Publish(
	ctx context.Context,
	input layoutapp.PublishStoreInput,
) (layoutapp.PublishedPresentation, error) {
	if !store.commandUsable(ctx) {
		return layoutapp.PublishedPresentation{}, layoutUnavailable()
	}
	var document []byte
	err := store.queryRow(
		ctx, publishSQL, input.ID, input.ExpectedRowRevision,
		input.SourceDefinitionSHA256[:], input.CompilerVersion,
		input.CatalogContract, string(input.CatalogJSON), input.CatalogSHA256[:],
		input.PresentationContract, string(input.PresentationJSON),
		input.PresentationSHA256[:], input.Now,
	).Scan(&document)
	if err != nil {
		return layoutapp.PublishedPresentation{}, mapLayoutError(err)
	}
	return decodePublishedPresentation(document)
}
