package aiworkspace

import (
	"context"

	catalogai "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/aiworkspace"
)

func (owner *Owner) List(ctx context.Context, query catalogai.ListQuery,
) ([]catalogai.Item, error) {
	kind := normalizedKind(query.Type)
	if !owner.allowed(ctx) {
		return nil, catalogai.ErrUnavailable
	}
	if query.Limit < 1 || query.Limit > 200 ||
		query.Type != "" && kind == "" {
		return nil, catalogai.ErrInvalid
	}
	if query.Tag != "" {
		return []catalogai.Item{}, nil
	}
	return owner.readItems(ctx, listSQL, kind, query.Limit)
}

func (owner *Owner) readItems(ctx context.Context, statement string,
	arguments ...any,
) ([]catalogai.Item, error) {
	resultRows, err := owner.query(ctx, statement, arguments...)
	if err != nil {
		return nil, catalogai.ErrUnavailable
	}
	defer resultRows.Close()
	result := make([]catalogai.Item, 0)
	for resultRows.Next() {
		item, scanErr := scanMenuItem(resultRows)
		projection, valid := item.projection()
		if scanErr != nil || !valid {
			return nil, catalogai.ErrUnavailable
		}
		result = append(result, projection)
	}
	if resultRows.Err() != nil {
		return nil, catalogai.ErrUnavailable
	}
	return result, nil
}

func scanMenuItem(source interface{ Scan(...any) error }) (menuItem, error) {
	var item menuItem
	err := source.Scan(&item.ItemID, &item.PackageID, &item.TargetID, &item.Kind,
		&item.Label, &item.Version, &item.UpdatedAt, &item.Score)
	return item, err
}
