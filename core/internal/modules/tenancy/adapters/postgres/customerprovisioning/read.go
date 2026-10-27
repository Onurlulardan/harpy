package customerprovisioning

import (
	"context"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/google/uuid"
)

func (store *Store) Read(ctx context.Context, operationID, subjectID uuid.UUID) (
	provisioning.Operation, error,
) {
	if !store.usable(ctx) || operationID == uuid.Nil || subjectID == uuid.Nil {
		return provisioning.Operation{}, provisioning.ErrInvalid
	}
	return scanOperation(
		store.queryRow(ctx, getSQL, operationID, subjectID),
		store.notFound,
	)
}
