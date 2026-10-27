package profilepostgres

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

func (storage *store) loadReadResource(
	ctx context.Context,
	bound *boundStore,
	contract profileWriteContract,
	resource identitydomain.ProfileResource,
	id uuid.UUID,
) (profileWriteResource, error) {
	if bound == nil || !contract.valid(resource) || id == uuid.Nil {
		return profileWriteResource{}, identitydomain.ErrStorage
	}
	var payload []byte
	if bound.database.QueryRow(
		ctx, profileDocumentSQL, string(resource), id,
	).Scan(&payload) != nil {
		return profileWriteResource{}, identitydomain.ErrStorage
	}
	current, err := decodeContractResource(ctx, bound, contract, resource, payload)
	if err != nil {
		return profileWriteResource{}, identitydomain.ErrStorage
	}
	return current, nil
}
