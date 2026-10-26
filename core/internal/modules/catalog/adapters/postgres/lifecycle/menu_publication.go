package lifecycle

import (
	"context"
	_ "embed"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

const menuDefinitionID = "018f6000-0000-7000-8000-000000000001"

//go:embed sql/tenant/ensure_menu_definition.sql
var ensureMenuDefinitionSQL string

//go:embed sql/tenant/lock_menu_definition.sql
var lockMenuDefinitionSQL string

//go:embed sql/tenant/read_current_menu_items.sql
var readCurrentMenuItemsSQL string

//go:embed sql/tenant/create_menu_version.sql
var createMenuVersionSQL string

//go:embed sql/tenant/copy_menu_items.sql
var copyMenuItemsSQL string

//go:embed sql/tenant/insert_menu_item.sql
var insertMenuItemSQL string

//go:embed sql/tenant/activate_menu_version.sql
var activateMenuVersionSQL string

type menuRecord struct {
	PackageID string
	Pin       catalogdomain.MenuPin
}

func (storage *store) publishOperationMenu(
	ctx context.Context, operation catalogdomain.Operation,
) error {
	if _, err := storage.execute(ctx, ensureMenuDefinitionSQL, menuDefinitionID, operation.StartedAt); err != nil {
		return unavailable()
	}
	var revision int64
	var activeVersion uuid.NullUUID
	if err := storage.queryRow(ctx, lockMenuDefinitionSQL, menuDefinitionID).Scan(
		&revision, &activeVersion); err != nil || revision < 1 {
		return unavailable()
	}
	current, err := storage.readCurrentMenu(ctx)
	if err != nil {
		return err
	}
	final := make([]menuRecord, 0, len(current)+len(operation.Plan.Menu))
	for _, item := range current {
		if item.PackageID != operation.Plan.PackageID.String() {
			final = append(final, item)
		}
	}
	for _, pin := range operation.Plan.Menu {
		final = append(final, menuRecord{PackageID: operation.Plan.PackageID.String(), Pin: pin})
	}
	contentSHA256 := menuDigest(final)
	if _, err = storage.execute(ctx, createMenuVersionSQL, operation.ID.String(),
		menuDefinitionID, revision+1, contentSHA256, operation.StartedAt); err != nil {
		return unavailable()
	}
	if activeVersion.Valid {
		if _, err = storage.execute(ctx, copyMenuItemsSQL, operation.ID.String(),
			activeVersion.UUID, operation.Plan.PackageID.String()); err != nil {
			return unavailable()
		}
	}
	for _, pin := range operation.Plan.Menu {
		if err = storage.insertMenuPin(ctx, operation.ID.String(), operation.Plan.PackageID.String(), pin); err != nil {
			return err
		}
	}
	rows, err := storage.execute(ctx, activateMenuVersionSQL,
		menuDefinitionID, operation.ID.String(), revision+1, operation.StartedAt, revision)
	if err != nil || rows != 1 {
		return unavailable()
	}
	return nil
}

func (storage *store) insertMenuPin(
	ctx context.Context, versionID, packageID string, pin catalogdomain.MenuPin,
) error {
	var parentID, targetID any
	if pin.ParentID != nil {
		parentID = pin.ParentID.String()
	}
	if pin.TargetID != nil {
		targetID = pin.TargetID.String()
	}
	var externalURL any
	if pin.URL != "" {
		externalURL = pin.URL
	}
	if _, err := storage.execute(ctx, insertMenuItemSQL, versionID, pin.ItemID.String(),
		parentID, packageID, pin.Kind, targetID, externalURL, pin.Label, pin.Order); err != nil {
		return unavailable()
	}
	return nil
}
