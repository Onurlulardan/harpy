package profilepostgres

import (
	"context"
	_ "embed"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/read_contract.sql
var profileReadContractSQL string

func (storage *store) loadReadContract(
	ctx context.Context,
	bound *boundStore,
	resource identitydomain.ProfileResource,
) (profileWriteContract, error) {
	return loadProfileContract(ctx, bound, profileReadContractSQL, resource)
}
