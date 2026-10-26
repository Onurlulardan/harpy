package application

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func (application *LayoutApplication) Get(
	ctx context.Context,
	id string,
) (layoutapi.Revision, error) {
	parsed, err := uuid.Parse(id)
	if !application.usable(ctx) || err != nil || !stableUUID(parsed) ||
		parsed.String() != id {
		return layoutapi.Revision{}, layoutapi.ErrInvalid
	}
	revision, err := application.store.Get(ctx, parsed)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	if revision.ID != id {
		return layoutapi.Revision{}, layoutapi.ErrUnavailable
	}
	return cloneRevision(revision), nil
}
