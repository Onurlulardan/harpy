package menu

import (
	"context"

	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func (storage *store) listOverrides(ctx context.Context, versionID string) ([]menudomain.Override, error) {
	resultRows, err := storage.query(ctx, listOverridesSQL, versionID)
	if err != nil {
		return nil, unavailable()
	}
	defer resultRows.Close()
	result := make([]menudomain.Override, 0)
	for resultRows.Next() {
		var itemRaw string
		var label *string
		var order *int64
		var visible *bool
		if err = resultRows.Scan(&itemRaw, &label, &order, &visible); err != nil {
			return nil, unavailable()
		}
		itemID, parseErr := packagecontract.ParseMenuItemID(itemRaw)
		if parseErr != nil || order != nil && (*order < 0 || *order > 100000) {
			return nil, unavailable()
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
		result = append(result, override)
	}
	if resultRows.Err() != nil {
		return nil, unavailable()
	}
	return result, nil
}
