package application

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func (application *LayoutApplication) Delete(
	ctx context.Context,
	input layoutapi.DeleteInput,
) error {
	id, err := uuid.Parse(input.ID)
	if !application.usable(ctx) || err != nil || !stableUUID(id) ||
		id.String() != input.ID ||
		input.ExpectedRowRevision < 1 {
		return layoutapi.ErrInvalid
	}
	return application.store.Delete(ctx, id, input.ExpectedRowRevision)
}
