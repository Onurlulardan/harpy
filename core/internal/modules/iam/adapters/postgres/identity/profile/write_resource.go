package profilepostgres

import (
	"context"
	_ "embed"
	"encoding/json"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

//go:embed sql/tenant/write_resource.sql
var profileWriteResourceSQL string

type profileWriteResource struct {
	actualSchemaID uuid.UUID
	actualSchema   profileapp.Schema
	revision       int64
	document       profileapp.Document
}

func (storage *store) loadWriteResource(
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
		ctx, profileWriteResourceSQL, string(resource), id,
	).Scan(&payload) != nil {
		return profileWriteResource{}, identitydomain.ErrStorage
	}
	current, err := decodeContractResource(ctx, bound, contract, resource, payload)
	if err != nil {
		return profileWriteResource{}, identitydomain.ErrStorage
	}
	var wire profileDocumentWire
	if json.Unmarshal(payload, &wire) != nil || wire.Revision < 1 {
		return profileWriteResource{}, identitydomain.ErrStorage
	}
	current.revision = wire.Revision
	return current, nil
}
