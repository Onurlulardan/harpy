package application

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func (application *LayoutApplication) Update(
	ctx context.Context,
	input layoutapi.UpdateInput,
) (layoutapi.Revision, error) {
	id, parseErr := uuid.Parse(input.ID)
	if !application.usable(ctx) || parseErr != nil || !stableUUID(id) ||
		id.String() != input.ID ||
		input.ExpectedRowRevision < 1 {
		return layoutapi.Revision{}, layoutapi.ErrInvalid
	}
	_, document, digest, err := encodeDefinition(input.Definition)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	now, err := application.currentTime()
	if err != nil {
		return layoutapi.Revision{}, err
	}
	revision, err := application.store.Update(ctx, UpdateStoreInput{
		ID: id, ExpectedRowRevision: input.ExpectedRowRevision,
		DefinitionJSON: string(document), DefinitionSHA256: digest,
		Now: now,
	})
	if err != nil {
		return layoutapi.Revision{}, err
	}
	if revision.ID != input.ID || revision.State != layoutapi.RevisionDraft ||
		revision.DefinitionSHA256 != digest ||
		revision.RowRevision != input.ExpectedRowRevision+1 {
		return layoutapi.Revision{}, layoutapi.ErrUnavailable
	}
	return cloneRevision(revision), nil
}
