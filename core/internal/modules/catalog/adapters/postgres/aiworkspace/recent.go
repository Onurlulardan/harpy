package aiworkspace

import (
	"context"
	"time"

	catalogai "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/aiworkspace"
)

func (owner *Owner) Recent(ctx context.Context, query catalogai.RecentQuery,
) ([]catalogai.Change, error) {
	if !owner.allowed(ctx) {
		return nil, catalogai.ErrUnavailable
	}
	if query.Since <= 0 || query.Since > 365*24*time.Hour ||
		query.Limit < 1 || query.Limit > 50 || len(query.Types) > 16 {
		return nil, catalogai.ErrInvalid
	}
	kinds := make([]string, 0, len(query.Types))
	for _, candidate := range query.Types {
		kind := normalizedKind(candidate)
		if kind == "" {
			return nil, catalogai.ErrInvalid
		}
		kinds = append(kinds, kind)
	}
	cutoff := time.Now().UTC().Add(-query.Since)
	items, err := owner.readItems(ctx, recentSQL, cutoff, kinds, query.Limit)
	if err != nil {
		return nil, err
	}
	result := make([]catalogai.Change, 0, len(items))
	for _, item := range items {
		change := "updated"
		result = append(result, catalogai.Change{Path: item.Path, Type: item.Type,
			ChangeType: change, ChangedAt: item.UpdatedAt,
			Requirements: append([]catalogai.Requirement(nil), item.Requirements...)})
	}
	return result, nil
}
