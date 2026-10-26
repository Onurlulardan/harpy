package billingv1

import (
	"encoding/json"
	"errors"
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type errorWire struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	document, err := json.Marshal(value)
	if err != nil {
		status, document = http.StatusServiceUnavailable,
			[]byte(`{"error":"Billing service unavailable"}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(document)
}

func (handler *Handler) fail(
	w http.ResponseWriter,
	routeID string,
	err error,
) {
	status, message, code := failureDetails(routeID, err)
	writeJSON(w, status, errorWire{Error: message, Code: code})
}

func failureDetails(routeID string, err error) (int, string, string) {
	if status, code, _, valid := billingapi.QueryErrorDetails(err); valid {
		return status, "Query could not be processed", code
	}
	switch {
	case errors.Is(err, billingapi.ErrAuthentication):
		return 401, "Authentication failed", ""
	case errors.Is(err, billingapi.ErrForbidden):
		return 403, "Operation is forbidden", ""
	case errors.Is(err, billingapi.ErrNotFound):
		return 404, "Resource not found", ""
	case errors.Is(err, billingapi.ErrTransitionInvalid):
		return 422, "Subscription transition is invalid", "BILLING_TRANSITION_INVALID"
	case errors.Is(err, ErrUnsupported):
		return 422, "Mutation is not supported by the immutable Billing model", "BILLING_MUTATION_UNSUPPORTED"
	case errors.Is(err, billingapi.ErrRevisionStale):
		return 409, "Billing resource revision is stale", "BILLING_REVISION_STALE"
	case errors.Is(err, billingapi.ErrImmutableConflict):
		return 409, "Billing resource conflicts", "BILLING_IMMUTABLE_CONFLICT"
	case errors.Is(err, billingapi.ErrConflict):
		if routeID == "billing.route.apply-coupon" {
			return 409, "Discount was already redeemed", "DISCOUNT_ALREADY_REDEEMED"
		}
		return 409, "Billing operation conflicts", ""
	case errors.Is(err, ErrRouteUnavailable):
		return 403, "This feature is only available in SaaS deployment mode", "SAAS_ONLY_FEATURE"
	case errors.Is(err, billingapi.ErrInvalid), errors.Is(err, ErrRequestInvalid),
		errors.Is(err, ErrSnapshotRequired), errors.Is(err, ErrPlanChangeTarget):
		return 400, "Invalid request", ""
	case errors.Is(err, billingapi.ErrAuthorityUnavailable),
		errors.Is(err, billingapi.ErrUnavailable), errors.Is(err, ErrRetirementEvidence),
		errors.Is(err, ErrAvailabilityEvidence),
		errors.Is(err, ErrPaginationEvidence):
		return 503, "Billing service unavailable", ""
	default:
		return 503, "Billing service unavailable", ""
	}
}
