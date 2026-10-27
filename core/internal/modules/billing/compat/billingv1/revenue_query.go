package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

const revenueReportDocument = `{
  "where": null,
  "aggregate": {
    "group_by": ["currency", "currency_registry_version"],
    "measures": [
      {"name":"monthly_minor_units","function":"sum","field":"monthly_minor_units"},
      {"name":"annual_minor_units","function":"sum","field":"annual_minor_units"}
    ],
    "having": null,
    "sort": [
      {"group":"currency","direction":"asc"},
      {"group":"currency_registry_version","direction":"asc"}
    ],
    "limit": 500
  }
}`

func (handler *Handler) revenueReportQuery(
	access billingapi.Access,
) (billingapi.Query, error) {
	request, err := handler.query.DecodePOST([]byte(revenueReportDocument))
	return billingapi.Query{
		Access: access, Request: request, Transport: http.MethodPost,
	}, err
}
