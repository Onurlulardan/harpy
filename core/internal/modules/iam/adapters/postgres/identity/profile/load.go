package profilepostgres

import (
	"context"
	"crypto/sha256"
	_ "embed"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/historical.sql
var profileHistoricalSQL string

//go:embed sql/tenant/document.sql
var profileDocumentSQL string

func (storage *store) Load(
	ctx context.Context, scope tenantScope, resource identitydomain.ProfileResource,
	id identitydomain.IdentityID,
) (identitydomain.ProfileDocument, error) {
	if !scopeUsable(ctx, scope, snapshotMode) || !resource.Valid() ||
		id == (identitydomain.IdentityID{}) {
		return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identitydomain.ProfileDocument{}, err
	}
	contract, err := storage.loadReadContract(ctx, bound, resource)
	if err != nil {
		return identitydomain.ProfileDocument{}, identitydomain.ErrStorage
	}
	current, err := storage.loadReadResource(
		ctx, bound, contract, resource, databaseID(id),
	)
	if err != nil {
		return identitydomain.ProfileDocument{}, err
	}
	_, projection, err := profileSourceRead(contract, current)
	if err != nil {
		return identitydomain.ProfileDocument{}, identitydomain.ErrStorage
	}
	digest := sha256.Sum256(projection)
	return identitydomain.NewProfileDocument(
		resource, identitydomain.IdentityID(contract.source.ID()), projection, digest,
	)
}
