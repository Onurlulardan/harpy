package billing

import (
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	billingreports "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/reports"
)

func recordPage[T any](
	items []T,
	result billingreports.RecordResult,
	query billingapi.Query,
) billingapi.Page[T] {
	var next *string
	if result.HasMore() {
		value := result.NextCursor()
		next = &value
	}
	return billingapi.Page[T]{
		Items: items,
		Pagination: billingapi.Pagination{
			Next: next, NextTransport: reportTransport(query.Transport),
			Size: query.Request.Page().Size(), Total: result.ExactCount(),
		},
		Meta: map[string]any{},
	}
}
