package iam

import (
	"context"

	"github.com/google/uuid"
)

func systemTransferReadInputValid(
	ctx context.Context,
	resource SystemTransferResource,
	transferID uuid.UUID,
) bool {
	return ctx != nil && ctx.Err() == nil && resource.Valid() &&
		transferID != uuid.Nil && transferID.Version() == 7 &&
		transferID.Variant() == uuid.RFC4122
}
