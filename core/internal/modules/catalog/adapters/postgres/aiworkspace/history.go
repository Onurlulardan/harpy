package aiworkspace

import (
	"context"

	catalogai "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/aiworkspace"
)

func (owner *Owner) History(ctx context.Context, path string, limit int,
) (catalogai.History, error) {
	kind, identity, err := parsePath(path)
	if !owner.allowed(ctx) {
		return catalogai.History{}, catalogai.ErrUnavailable
	}
	if err != nil || limit < 1 || limit > 100 {
		return catalogai.History{}, catalogai.ErrInvalid
	}
	current, err := owner.readOne(ctx, describeSQL, identity, kind)
	if err != nil {
		return catalogai.History{}, err
	}
	resultRows, err := owner.query(ctx, historySQL, current.ItemID, limit)
	if err != nil {
		return catalogai.History{}, catalogai.ErrUnavailable
	}
	defer resultRows.Close()
	requirements := ownerRequirements(current.ItemID)
	result := catalogai.History{Requirements: requirements}
	for resultRows.Next() {
		var version int
		var value catalogai.Version
		if resultRows.Scan(&version, &value.PublishedAt, &value.IsActive) != nil || version < 1 {
			return catalogai.History{}, catalogai.ErrUnavailable
		}
		value.Version = version
		value.PublishedAt = value.PublishedAt.UTC()
		value.Requirements = append([]catalogai.Requirement(nil), requirements...)
		result.Items = append(result.Items, value)
	}
	if resultRows.Err() != nil {
		return catalogai.History{}, catalogai.ErrUnavailable
	}
	return result, nil
}
