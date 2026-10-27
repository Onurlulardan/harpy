package application

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func (application *LayoutApplication) Create(
	ctx context.Context,
	input layoutapi.CreateInput,
) (layoutapi.Revision, error) {
	if !application.usable(ctx) || !input.Resource.Valid() {
		return layoutapi.Revision{}, layoutapi.ErrInvalid
	}
	id, err := createRevisionID(input.ID, application.ids)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	_, document, digest, err := encodeDefinition(input.Definition)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	now, err := application.currentTime()
	if err != nil {
		return layoutapi.Revision{}, err
	}
	revision, err := application.store.Create(ctx, CreateStoreInput{
		ID: id, Resource: input.Resource, DefinitionJSON: string(document),
		DefinitionSHA256: digest, Now: now,
	})
	if err != nil {
		return layoutapi.Revision{}, err
	}
	if revision.ID != id.String() || revision.Resource != input.Resource ||
		revision.Version < 1 || !revision.State.Valid() ||
		revision.DefinitionSHA256 != digest || revision.RowRevision < 1 {
		return layoutapi.Revision{}, layoutapi.ErrUnavailable
	}
	return cloneRevision(revision), nil
}

func createRevisionID(wire string, source IDSource) (uuid.UUID, error) {
	if wire == "" {
		value, err := source.NewUUID()
		if err != nil || !revisionUUIDV7(value) {
			return uuid.Nil, layoutapi.ErrUnavailable
		}
		return value, nil
	}
	value, err := uuid.Parse(wire)
	if err != nil || value.String() != wire || !revisionUUIDV7(value) {
		return uuid.Nil, layoutapi.ErrInvalid
	}
	return value, nil
}

func revisionUUIDV7(value uuid.UUID) bool {
	return stableUUID(value) && value.Version() == 7
}
