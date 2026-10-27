package billing

import (
	"encoding/json"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type queryResult interface {
	Items() []json.RawMessage
	HasMore() bool
	NextCursor() string
	NextTransport() string
	ExactCount() *int64
}

func toQueryPage(result queryResult) billingapi.QueryPage {
	items := result.Items()
	var next *string
	if result.HasMore() {
		value := result.NextCursor()
		next = &value
	}
	return billingapi.QueryPage{
		Items: items,
		Pagination: billingapi.Pagination{
			Next: next, NextTransport: result.NextTransport(),
			Size: len(items), Total: result.ExactCount(),
		},
		Meta: map[string]any{},
	}
}
