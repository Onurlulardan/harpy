package menu

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ownerTargetAuthorizer struct{ owners componentowner.Directory }

func NewOwnerTargetAuthorizer(owners componentowner.Directory) (TargetAuthorizer, error) {
	if owners == nil {
		return nil, fmt.Errorf("catalog menu target owners are required")
	}
	return &ownerTargetAuthorizer{owners: owners}, nil
}

func (authorizer *ownerTargetAuthorizer) Allows(ctx context.Context,
	scope platformtx.TenantScope, kind packagecontract.MenuKind,
	target packagecontract.ComponentID,
) (bool, error) {
	ownerName, ok := targetOwner(kind)
	if authorizer == nil || authorizer.owners == nil || !ok || !target.Valid() {
		return false, fmt.Errorf("catalog menu target is invalid")
	}
	owner, err := authorizer.owners.Resolve(ownerName)
	if err != nil {
		return false, err
	}
	menuOwner, ok := owner.(componentowner.MenuAuthorizer)
	if !ok || menuOwner == nil {
		return false, fmt.Errorf("catalog menu owner is unavailable")
	}
	return menuOwner.AllowsMenuTarget(ctx, scope, string(kind), target.String())
}

func targetOwner(kind packagecontract.MenuKind) (string, bool) {
	switch kind {
	case packagecontract.MenuForm:
		return "forms", true
	case packagecontract.MenuWorkflow:
		return "workflow", true
	case packagecontract.MenuDashboard, packagecontract.MenuEndpoint:
		return "analytics", true
	default:
		return "", false
	}
}

var _ TargetAuthorizer = (*ownerTargetAuthorizer)(nil)
