package menu

import (
	"context"

	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Replace(
	ctx context.Context, scope platformtx.TenantScope, request ReplaceRequest,
) (Snapshot, error) {
	if request.ExpectedRevision < 1 || !idempotencyPattern.MatchString(request.IdempotencyKey) ||
		len(request.Items) > 2048 {
		return Snapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeCommand, operationUpdate, permissionManage)
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	versionID, idErr := application.ids.New()
	if !canonicalTime(now) || idErr != nil || versionID == uuid.Nil || versionID.Version() != 7 {
		return Snapshot{}, ErrUnavailable
	}
	items, err := application.replaceItems(request.Items)
	if err != nil {
		return Snapshot{}, err
	}
	source, err := store.Replace(ctx, ReplaceMutation{VersionID: versionID,
		ExpectedRevision: request.ExpectedRevision, Items: items, OccurredAt: now,
		ContentSHA256: replaceDigest(items)})
	if err != nil {
		return Snapshot{}, err
	}
	return snapshotFromSource(source)
}

func (application *Application) replaceItems(raw []ReplaceItem) ([]menudomain.Item, error) {
	packageID, err := packagecontract.ParsePackageID(TenantAuthoredPackageID)
	if err != nil {
		return nil, ErrUnavailable
	}
	resolved := make([]packagecontract.MenuItemID, len(raw))
	ids := map[string]packagecontract.MenuItemID{}
	for index, entry := range raw {
		itemID, idErr := application.menuItemID(entry.ItemID)
		if idErr != nil {
			return nil, idErr
		}
		resolved[index] = itemID
		ids[itemID.String()] = itemID
		if entry.ItemID != "" {
			ids[entry.ItemID] = itemID
		}
	}
	items := make([]menudomain.Item, 0, len(raw))
	for index, entry := range raw {
		item, itemErr := replaceItem(entry, packageID, resolved[index], ids)
		if itemErr != nil {
			return nil, itemErr
		}
		items = append(items, item)
	}
	return items, nil
}

func replaceItem(
	entry ReplaceItem, packageID packagecontract.PackageID,
	itemID packagecontract.MenuItemID, ids map[string]packagecontract.MenuItemID,
) (menudomain.Item, error) {
	item := menudomain.Item{MenuItem: packagecontract.MenuItem{ID: itemID,
		Kind: packagecontract.MenuKind(entry.Kind), Label: entry.Label, Order: entry.Order},
		PackageID: packageID, Visible: entry.Visible}
	if entry.ParentID != nil && *entry.ParentID != "" {
		parent, found := ids[*entry.ParentID]
		if !found {
			return menudomain.Item{}, ErrInvalid
		}
		item.ParentID = &parent
	}
	if entry.TargetID != nil && *entry.TargetID != "" {
		target, targetErr := packagecontract.ParseComponentID(*entry.TargetID)
		if targetErr != nil {
			return menudomain.Item{}, ErrInvalid
		}
		item.TargetID = &target
	}
	if !validReplaceItem(item) {
		return menudomain.Item{}, ErrInvalid
	}
	return item, nil
}

func (application *Application) menuItemID(raw string) (packagecontract.MenuItemID, error) {
	if parsed, err := packagecontract.ParseMenuItemID(raw); err == nil {
		return parsed, nil
	}
	id, err := application.ids.New()
	if err != nil || id == uuid.Nil || id.Version() != 7 {
		return packagecontract.MenuItemID{}, ErrUnavailable
	}
	return packagecontract.ParseMenuItemID(id.String())
}

func validReplaceItem(item menudomain.Item) bool {
	if item.Label == "" || len(item.Label) > 128 || item.Order < 0 || item.Order > 100000 {
		return false
	}
	switch item.Kind {
	case packagecontract.MenuGroup:
		return item.TargetID == nil
	case packagecontract.MenuForm, packagecontract.MenuWorkflow,
		packagecontract.MenuDashboard, packagecontract.MenuEndpoint:
		return item.TargetID != nil
	default:
		return false
	}
}

func snapshotFromSource(source Source) (Snapshot, error) {
	if !source.Version.ID.Valid() || source.Version.Number < 1 || source.Version.Revision < 1 {
		return Snapshot{}, ErrUnavailable
	}
	versionID, err := uuid.Parse(source.Version.ID.String())
	if err != nil {
		return Snapshot{}, ErrUnavailable
	}
	items, err := menudomain.ApplyOverrides(source.Version.Items, source.Overrides)
	if err != nil {
		return Snapshot{}, ErrUnavailable
	}
	return Snapshot{VersionID: versionID, Number: source.Version.Number,
		Revision: source.Version.Revision, Items: items}, nil
}
