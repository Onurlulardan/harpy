package application

import (
	"context"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func (application *PresentationApplication) Publish(
	ctx context.Context,
	input layoutapi.PublishPresentationInput,
) (layoutapi.Revision, error) {
	id, parseErr := uuid.Parse(input.RevisionID)
	if !application.usable(ctx) || parseErr != nil || !stableUUID(id) ||
		id.String() != input.RevisionID || input.ExpectedRowRevision < 1 ||
		input.ExpectedRowRevision == maximumRowRevision ||
		input.Catalog.Validate() != nil {
		return layoutapi.Revision{}, layoutapi.ErrInvalid
	}
	revision, err := application.store.Get(ctx, id)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	if revision.ID != input.RevisionID || !publishRevisionMatches(
		revision, input.ExpectedRowRevision,
	) {
		return layoutapi.Revision{}, layoutapi.ErrConflict
	}
	compiled, err := compilePresentation(revision, input.Catalog)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	catalogJSON, err := input.Catalog.CanonicalJSON()
	if err != nil {
		return layoutapi.Revision{}, layoutapi.ErrInvalid
	}
	catalogSHA, err := input.Catalog.CanonicalSHA256()
	if err != nil {
		return layoutapi.Revision{}, layoutapi.ErrInvalid
	}
	now, err := application.currentTime()
	if err != nil {
		return layoutapi.Revision{}, err
	}
	published, err := application.store.Publish(ctx, PublishStoreInput{
		ID: id, ExpectedRowRevision: input.ExpectedRowRevision,
		SourceDefinitionSHA256: revision.DefinitionSHA256,
		CompilerVersion:        compiled.compilerVersion,
		CatalogContract:        presentation.FieldCatalogContract,
		CatalogJSON:            catalogJSON, CatalogSHA256: catalogSHA,
		PresentationContract: presentation.ContractVersion,
		PresentationJSON:     compiled.canonicalJSON,
		PresentationSHA256:   compiled.digest, Now: now,
	})
	if err != nil {
		return layoutapi.Revision{}, err
	}
	if published.Revision.ID != input.RevisionID ||
		published.Revision.RowRevision != input.ExpectedRowRevision+1 {
		return layoutapi.Revision{}, layoutapi.ErrUnavailable
	}
	if _, err := verifyPublishedPresentation(
		published, revision.Resource, input.Catalog,
	); err != nil {
		return layoutapi.Revision{}, err
	}
	return cloneRevision(published.Revision), nil
}

func publishRevisionMatches(revision layoutapi.Revision, expected int64) bool {
	switch revision.State {
	case layoutapi.RevisionDraft:
		return revision.RowRevision == expected
	case layoutapi.RevisionPublished:
		return revision.RowRevision == expected+1
	default:
		return false
	}
}

const maximumRowRevision = int64(^uint64(0) >> 1)
