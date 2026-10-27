package menu

import (
	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func scanOverride(source row) (menuapp.OverrideReceipt, error) {
	var itemRaw string
	var label *string
	var order *int64
	var visible *bool
	var revision int64
	if err := source.Scan(&itemRaw, &label, &order, &visible, &revision); err != nil {
		return menuapp.OverrideReceipt{}, err
	}
	itemID, err := packagecontract.ParseMenuItemID(itemRaw)
	if err != nil || revision < 1 || order != nil && (*order < 0 || *order > 100000) {
		return menuapp.OverrideReceipt{}, unavailable()
	}
	override := menudomain.Override{ItemID: itemID}
	if label != nil {
		override.Label = label
	}
	if order != nil {
		value := int(*order)
		override.Order = &value
	}
	if visible != nil {
		override.Visible = visible
	}
	return menuapp.OverrideReceipt{Override: override, Revision: uint64(revision)}, nil
}

func overrideEqual(left, right menudomain.Override) bool {
	return left.ItemID.String() == right.ItemID.String() &&
		stringEqual(left.Label, right.Label) && intEqual(left.Order, right.Order) &&
		boolEqual(left.Visible, right.Visible)
}

func stringEqual(left, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func intEqual(left, right *int) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func boolEqual(left, right *bool) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}
