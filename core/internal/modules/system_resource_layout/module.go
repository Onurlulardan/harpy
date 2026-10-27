package systemresourcelayout

import (
	"context"

	layoutpostgres "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/adapters/postgres"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ReferenceTarget = layoutapp.ReferenceTarget
type ReferenceEvidence = layoutapp.ReferenceEvidence
type ResourceKind = layoutapi.ResourceKind

const (
	ResourceUser         = layoutapi.ResourceUser
	ResourceOrganization = layoutapi.ResourceOrganization
	ResourceRole         = layoutapi.ResourceRole
)

var ErrInvalid = layoutapi.ErrInvalid

type References struct{}

func NewReferences() References {
	return References{}
}

func (References) FieldReferenceEvidence(
	ctx context.Context,
	scope platformtx.TenantScope,
	target ReferenceTarget,
) (ReferenceEvidence, error) {
	store, err := layoutpostgres.BindReferenceStore(scope)
	if err != nil {
		return ReferenceEvidence{}, err
	}
	return store.FieldReferenceEvidence(ctx, target)
}
