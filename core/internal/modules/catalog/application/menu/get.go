package menu

import (
	"context"

	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Get(
	ctx context.Context, scope platformtx.TenantScope,
) (Snapshot, error) {
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, operationGet, permissionRead)
	if err != nil {
		return Snapshot{}, err
	}
	source, err := store.Get(ctx)
	if err != nil || !source.Version.ID.Valid() || source.Version.Number < 1 || source.Version.Revision < 1 {
		return Snapshot{}, ErrUnavailable
	}
	items, err := menudomain.ApplyOverrides(source.Version.Items, source.Overrides)
	if err != nil {
		return Snapshot{}, ErrUnavailable
	}
	decisions := targetDecisions{allowed: map[string]bool{}}
	allowTenantAuthoredTargets(items, decisions.allowed)
	for _, item := range items {
		if item.TargetID == nil || item.Kind == packagecontract.MenuGroup ||
			item.Kind == packagecontract.MenuExternalLink ||
			decisions.allowed[item.TargetID.String()] {
			continue
		}
		allowed, decisionErr := application.targets.Allows(ctx, scope, item.Kind, *item.TargetID)
		if decisionErr != nil {
			return Snapshot{}, ErrUnavailable
		}
		decisions.allowed[item.TargetID.String()] = allowed
	}
	effective := menudomain.Effective(items, activateTenantAuthored(source.ActivePackages), decisions)
	versionID, err := uuid.Parse(source.Version.ID.String())
	if err != nil {
		return Snapshot{}, ErrUnavailable
	}
	return Snapshot{VersionID: versionID, Number: source.Version.Number,
		Revision: source.Version.Revision, Items: effective}, nil
}

type targetDecisions struct{ allowed map[string]bool }

func (value targetDecisions) Allows(
	_ packagecontract.MenuKind, id packagecontract.ComponentID,
) bool {
	return value.allowed[id.String()]
}
