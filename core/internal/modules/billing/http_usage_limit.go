package billing

import (
	"errors"
	"net/http"
	"strconv"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type usageLimitDetails struct {
	EntitlementKey string `json:"entitlement_key"`
	Current        string `json:"current"`
	Requested      string `json:"requested"`
	Limit          string `json:"limit"`
	PeriodStart    string `json:"period_start"`
	PeriodEnd      string `json:"period_end"`
	ResetPolicy    string `json:"reset_policy"`
}

type billingError struct {
	Code          string            `json:"code"`
	Message       string            `json:"message"`
	Issues        []runtimev1.Issue `json:"issues"`
	Details       usageLimitDetails `json:"details"`
	CorrelationID string            `json:"correlation_id"`
}

type billingErrorEnvelope struct {
	Success bool         `json:"success"`
	Error   billingError `json:"error"`
}

func (handler *Handler) writeUsageLimit(
	w http.ResponseWriter,
	r *http.Request,
	usage billingapi.UsageLimitExceeded,
) {
	details := usageLimitDetails{
		EntitlementKey: usage.EntitlementKey,
		Current:        strconv.FormatInt(usage.Current, 10),
		Requested:      strconv.FormatInt(usage.Requested, 10),
		Limit:          strconv.FormatInt(usage.Limit, 10),
		PeriodStart:    usage.PeriodStart, PeriodEnd: usage.PeriodEnd,
		ResetPolicy: usage.ResetPolicy,
	}
	envelope := billingErrorEnvelope{Success: false, Error: billingError{
		Code: "USAGE_LIMIT_EXCEEDED", Message: "Usage limit was exceeded",
		Issues: []runtimev1.Issue{}, Details: details,
		CorrelationID: correlationID(r).String(),
	}}
	if err := httpwire.WriteJSON(w, http.StatusPaymentRequired, correlationID(r), envelope, false); err != nil {
		handler.report(errors.New("billing usage-limit failure write failed"))
	}
}
