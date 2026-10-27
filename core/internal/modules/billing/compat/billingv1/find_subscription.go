package billingv1

import (
	"net/http"
	"net/url"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

func (handler *Handler) findSubscription(
	r *http.Request,
	access billingapi.Access,
	id string,
) (billingapi.SubscriptionView, error) {
	identifier, err := uuid.Parse(id)
	if err != nil || identifier == uuid.Nil || identifier.String() != id {
		return billingapi.SubscriptionView{}, ErrRequestInvalid
	}
	for _, field := range []string{"legacy_id", "id"} {
		query, decodeErr := handler.subscriptionLookupQuery(access, field, id)
		if decodeErr != nil {
			return billingapi.SubscriptionView{}, decodeErr
		}
		page, queryErr := handler.reports.ListSubscriptions(
			r.Context(), query,
		)
		if queryErr != nil {
			return billingapi.SubscriptionView{}, queryErr
		}
		if len(page.Items) > 1 || page.Pagination.Next != nil {
			return billingapi.SubscriptionView{}, ErrSnapshotRequired
		}
		if len(page.Items) == 1 {
			return materializeSubscription(page.Items[0])
		}
	}
	return billingapi.SubscriptionView{}, billingapi.ErrNotFound
}

func (handler *Handler) subscriptionLookupQuery(
	access billingapi.Access,
	field, id string,
) (billingapi.Query, error) {
	values := url.Values{
		"page[size]":       []string{"2"},
		"filter[0][field]": []string{field},
		"filter[0][op]":    []string{"eq"},
		"filter[0][value]": []string{id},
	}
	request, err := handler.query.DecodeGET(values.Encode())
	return billingapi.Query{
		Access: access, Request: request, Transport: http.MethodGet,
	}, err
}
