package billingv1

import (
	"context"
	"encoding/json"
	"net/url"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

const (
	legacyPageSize   = "500"
	maximumPageCount = 200
)

type catalogPageFetcher func(
	context.Context, billingapi.QueryRequest,
) (billingapi.QueryPage, error)

func (handler *Handler) collectCatalogPages(
	ctx context.Context,
	fetch catalogPageFetcher,
) ([]json.RawMessage, error) {
	items := make([]json.RawMessage, 0)
	cursor, seen := "", map[string]struct{}{}
	for pageIndex := 0; pageIndex < maximumPageCount; pageIndex++ {
		values := url.Values{"page[size]": []string{legacyPageSize}}
		if cursor != "" {
			values.Set("page[after]", cursor)
		}
		request, err := handler.query.DecodeGET(values.Encode())
		if err != nil {
			return nil, err
		}
		page, err := fetch(ctx, request)
		if err != nil {
			return nil, err
		}
		items = append(items, page.Items...)
		next, more, err := nextPage(page.Pagination, seen)
		if err != nil {
			return nil, err
		}
		if !more {
			return items, nil
		}
		cursor = next
	}
	return nil, ErrPaginationEvidence
}

func nextPage(
	pagination billingapi.Pagination,
	seen map[string]struct{},
) (string, bool, error) {
	if pagination.Next == nil {
		return "", false, nil
	}
	next := *pagination.Next
	if next == "" {
		return "", false, ErrPaginationEvidence
	}
	if _, exists := seen[next]; exists {
		return "", false, ErrPaginationEvidence
	}
	seen[next] = struct{}{}
	return next, true, nil
}
