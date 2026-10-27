package billingv1

import (
	"context"
	"net/http"
	"net/url"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) allSubscriptions(
	ctx context.Context,
	access billingapi.Access,
) ([]billingapi.SubscriptionView, error) {
	result := make([]billingapi.SubscriptionView, 0)
	cursor, seen := "", map[string]struct{}{}
	for pageIndex := 0; pageIndex < maximumPageCount; pageIndex++ {
		query, err := handler.subscriptionReportQuery(access, cursor)
		if err != nil {
			return nil, err
		}
		page, err := handler.reports.ListSubscriptions(ctx, query)
		if err != nil {
			return nil, err
		}
		for _, item := range page.Items {
			view, materializeErr := materializeSubscription(item)
			if materializeErr != nil {
				return nil, materializeErr
			}
			result = append(result, view)
		}
		next, more, err := nextPage(page.Pagination, seen)
		if err != nil {
			return nil, err
		}
		if !more {
			return result, nil
		}
		cursor = next
	}
	return nil, ErrPaginationEvidence
}

func (handler *Handler) subscriptionReportQuery(
	access billingapi.Access,
	after string,
) (billingapi.Query, error) {
	values := url.Values{
		"page[size]":         []string{legacyPageSize},
		"sort[0][field]":     []string{"created_at"},
		"sort[0][direction]": []string{"desc"},
	}
	if after != "" {
		values.Set("page[after]", after)
	}
	request, err := handler.query.DecodeGET(values.Encode())
	return billingapi.Query{
		Access: access, Request: request, Transport: http.MethodGet,
	}, err
}
