package billingv1

import (
	"encoding/json"
	"fmt"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func materializeLegacyEvent(
	view billingapi.LegacyEventView,
) (billingapi.LegacyEventView, error) {
	document, err := json.Marshal(view)
	if err != nil {
		return billingapi.LegacyEventView{}, fmt.Errorf(
			"marshal Billing V1 event query result: %w", err,
		)
	}
	var result billingapi.LegacyEventView
	if err := json.Unmarshal(document, &result); err != nil {
		return billingapi.LegacyEventView{}, fmt.Errorf(
			"decode Billing V1 event query result: %w", err,
		)
	}
	return result, nil
}
