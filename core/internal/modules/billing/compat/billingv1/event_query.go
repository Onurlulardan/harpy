package billingv1

import (
	"net/http"
	"net/url"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

func (handler *Handler) eventReportQuery(
	access billingapi.Access,
	tenantID string,
) (billingapi.Query, error) {
	values := url.Values{
		"page[size]":         []string{"100"},
		"sort[0][field]":     []string{"created_at"},
		"sort[0][direction]": []string{"desc"},
	}
	if tenantID != "" {
		parsed, err := uuid.Parse(tenantID)
		if err != nil || parsed == uuid.Nil || parsed.String() != tenantID {
			return billingapi.Query{}, ErrRequestInvalid
		}
		values.Set("filter[0][field]", "tenant_id")
		values.Set("filter[0][op]", "eq")
		values.Set("filter[0][value]", parsed.String())
	}
	request, err := handler.query.DecodeGET(values.Encode())
	return billingapi.Query{
		Access: access, Request: request, Transport: http.MethodGet,
	}, err
}
