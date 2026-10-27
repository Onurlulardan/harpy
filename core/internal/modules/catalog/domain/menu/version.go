package menu

import (
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

type Override struct {
	ItemID  packagecontract.MenuItemID
	Label   *string
	Order   *int
	Visible *bool
}

type Item struct {
	packagecontract.MenuItem
	PackageID packagecontract.PackageID
	Visible   bool
}

type Version struct {
	ID       packagecontract.StableID
	Number   uint64
	Revision uint64
	Items    []Item
}

func NewVersion(id packagecontract.StableID, number, expectedRevision uint64, items []Item) (Version, error) {
	if !id.Valid() || number == 0 || expectedRevision == 0 || len(items) > 2048 {
		return Version{}, ErrInvalid
	}
	seen := make(map[string]bool, len(items))
	copyItems := append([]Item(nil), items...)
	for _, item := range copyItems {
		if !item.ID.Valid() || !item.PackageID.Valid() || seen[item.ID.String()] {
			return Version{}, ErrInvalid
		}
		seen[item.ID.String()] = true
	}
	sort.Slice(copyItems, func(i, j int) bool {
		if copyItems[i].Order != copyItems[j].Order {
			return copyItems[i].Order < copyItems[j].Order
		}
		return copyItems[i].ID.String() < copyItems[j].ID.String()
	})
	return Version{ID: id, Number: number,
		Revision: expectedRevision + 1, Items: copyItems}, nil
}

func ApplyOverrides(items []Item, overrides []Override) ([]Item, error) {
	byID := make(map[string]int, len(items))
	result := append([]Item(nil), items...)
	for index, item := range result {
		byID[item.ID.String()] = index
	}
	for _, override := range overrides {
		index, found := byID[override.ItemID.String()]
		if !found {
			return nil, ErrInvalid
		}
		if override.Label != nil {
			if *override.Label == "" || len(*override.Label) > 128 {
				return nil, ErrInvalid
			}
			result[index].Label = *override.Label
		}
		if override.Order != nil {
			if *override.Order < 0 || *override.Order > 100000 {
				return nil, ErrInvalid
			}
			result[index].Order = *override.Order
		}
		if override.Visible != nil {
			result[index].Visible = *override.Visible
		}
	}
	return result, nil
}
