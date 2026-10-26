package menu

import (
	"context"
	"net/url"

	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func (storage *store) listItems(ctx context.Context, versionID string) ([]menudomain.Item, error) {
	resultRows, err := storage.query(ctx, listItemsSQL, versionID)
	if err != nil {
		return nil, unavailable()
	}
	defer resultRows.Close()
	items := make([]menudomain.Item, 0)
	identities := map[string]bool{}
	for resultRows.Next() {
		item, scanErr := scanItem(resultRows)
		if scanErr != nil || identities[item.ID.String()] {
			return nil, unavailable()
		}
		identities[item.ID.String()] = true
		items = append(items, item)
	}
	if resultRows.Err() != nil || !validParents(items, identities) {
		return nil, unavailable()
	}
	return items, nil
}

func scanItem(source row) (menudomain.Item, error) {
	var itemRaw, packageRaw, kindRaw, label string
	var parentRaw, targetRaw, externalURL *string
	var order int
	if err := source.Scan(&itemRaw, &parentRaw, &packageRaw, &kindRaw,
		&targetRaw, &externalURL, &label, &order); err != nil {
		return menudomain.Item{}, err
	}
	itemID, itemErr := packagecontract.ParseMenuItemID(itemRaw)
	packageID, packageErr := packagecontract.ParsePackageID(packageRaw)
	item := menudomain.Item{MenuItem: packagecontract.MenuItem{ID: itemID,
		Kind: packagecontract.MenuKind(kindRaw), Label: label, Order: order,
		URL: menuString(externalURL)}, PackageID: packageID, Visible: true}
	if itemErr != nil || packageErr != nil || !decodeReferences(&item, parentRaw, targetRaw) ||
		!validItem(item) {
		return menudomain.Item{}, unavailable()
	}
	return item, nil
}

func decodeReferences(item *menudomain.Item, parentRaw, targetRaw *string) bool {
	if parentRaw != nil {
		parent, err := packagecontract.ParseMenuItemID(*parentRaw)
		if err != nil || parent.String() == item.ID.String() {
			return false
		}
		item.ParentID = &parent
	}
	if targetRaw != nil {
		target, err := packagecontract.ParseComponentID(*targetRaw)
		if err != nil {
			return false
		}
		item.TargetID = &target
	}
	return true
}

func menuString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validItem(item menudomain.Item) bool {
	if item.Label == "" || len(item.Label) > 128 || item.Order < 0 || item.Order > 100000 {
		return false
	}
	if item.Kind == packagecontract.MenuGroup {
		return item.TargetID == nil && item.URL == ""
	}
	if item.Kind == packagecontract.MenuExternalLink {
		parsed, err := url.ParseRequestURI(item.URL)
		return err == nil && parsed.Scheme == "https" && parsed.Host != "" && item.TargetID == nil
	}
	return item.TargetID != nil && item.URL == "" && (item.Kind == packagecontract.MenuForm ||
		item.Kind == packagecontract.MenuWorkflow || item.Kind == packagecontract.MenuDashboard ||
		item.Kind == packagecontract.MenuEndpoint)
}

func validParents(items []menudomain.Item, identities map[string]bool) bool {
	for _, item := range items {
		if item.ParentID != nil && !identities[item.ParentID.String()] {
			return false
		}
	}
	return true
}
