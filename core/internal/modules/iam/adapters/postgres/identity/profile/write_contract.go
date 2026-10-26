package profilepostgres

import (
	"context"
	_ "embed"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

//go:embed sql/tenant/write_contract.sql
var profileWriteContractSQL string

type profileWriteContract struct {
	planID   uuid.UUID
	source   profileapp.Schema
	target   profileapp.Schema
	mappings []profileapp.MigrationMapping
}

func (contract profileWriteContract) migrating() bool {
	return contract.planID != uuid.Nil
}

func (contract profileWriteContract) valid(
	resource identitydomain.ProfileResource,
) bool {
	if string(contract.source.Resource()) != string(resource) ||
		contract.source.State() != profileapp.SchemaActive {
		return false
	}
	if !contract.migrating() {
		return contract.target.ID() == uuid.Nil && len(contract.mappings) == 0
	}
	return contract.target.ID() != uuid.Nil &&
		contract.target.State() == profileapp.SchemaActive &&
		contract.target.Resource() == contract.source.Resource() &&
		contract.target.Version() == contract.source.Version()+1 &&
		len(contract.mappings) > 0
}

func (storage *store) loadWriteContract(
	ctx context.Context,
	bound *boundStore,
	resource identitydomain.ProfileResource,
) (profileWriteContract, error) {
	return loadProfileContract(ctx, bound, profileWriteContractSQL, resource)
}

func loadProfileContract(
	ctx context.Context,
	bound *boundStore,
	query string,
	resource identitydomain.ProfileResource,
) (profileWriteContract, error) {
	if bound == nil || !resource.Valid() {
		return profileWriteContract{}, identitydomain.ErrStorage
	}
	var payload []byte
	if err := bound.database.QueryRow(
		ctx, query, string(resource),
	).Scan(&payload); err != nil {
		return profileWriteContract{}, profileContractError(err)
	}
	contract, err := decodeProfileWriteContract(payload)
	if err != nil || !contract.valid(resource) {
		return profileWriteContract{}, identitydomain.ErrStorage
	}
	return contract, nil
}
