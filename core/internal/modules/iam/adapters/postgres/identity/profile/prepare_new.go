package profilepostgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

func (storage *store) PrepareNew(
	ctx context.Context,
	scope tenantScope,
	resource identitydomain.ProfileResource,
	values []byte,
) (identitydomain.ProfileDocument, error) {
	if !scopeUsable(ctx, scope, commandMode) || !resource.Valid() {
		return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identitydomain.ProfileDocument{}, identitydomain.ErrStorage
	}
	contract, err := storage.loadWriteContract(ctx, bound, resource)
	if err != nil {
		return identitydomain.ProfileDocument{}, err
	}
	document, err := profileapp.DecodeDocument(
		contract.source, values, nil, profileNormalizer{},
	)
	if err != nil {
		return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidCommand
	}
	if contract.migrating() {
		document, err = migrateProfileWrite(contract, document)
		if err != nil {
			return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidCommand
		}
	}
	return newProfileDocument(resource, document)
}
