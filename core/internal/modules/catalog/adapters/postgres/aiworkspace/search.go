package aiworkspace

import (
	"context"
	"strings"

	catalogai "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/aiworkspace"
)

func (owner *Owner) Search(ctx context.Context, query catalogai.SearchQuery,
) ([]catalogai.Item, error) {
	value := strings.TrimSpace(query.Query)
	if !owner.allowed(ctx) {
		return nil, catalogai.ErrUnavailable
	}
	if value == "" || query.Limit < 1 || query.Limit > 25 ||
		len(query.Types) > 16 {
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
	return owner.readItems(ctx, searchSQL, value, kinds, query.Limit)
}
