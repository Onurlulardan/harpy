package billing

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

func (handler *Handler) writeFailure(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	var usage billingapi.UsageLimitExceeded
	if errors.As(err, &usage) && usage.Valid() {
		handler.writeUsageLimit(w, r, usage)
		return
	}
	status, code, message, issues := billingFailure(err)
	if status >= 500 {
		handler.report(errors.New("billing HTTP operation failed"))
	}
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(
		code, message, correlationID(r).String(), issues,
	)
	if envelopeErr == nil {
		envelopeErr = httpwire.WriteJSON(w, status, correlationID(r), envelope, false)
	}
	if envelopeErr != nil {
		handler.report(errors.New("billing HTTP failure write failed"))
	}
}

func billingFailure(err error) (int, string, string, []runtimev1.Issue) {
	if status, code, issues, valid := billingapi.QueryErrorDetails(err); valid {
		return status, code, "Query could not be processed", issues
	}
	switch {
	case errors.Is(err, httpinput.ErrBodyTooLarge):
		return 413, "REQUEST_BODY_TOO_LARGE", "Request body is too large", nil
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return 415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported", nil
	case errors.Is(err, billingapi.ErrAuthentication):
		return 401, "AUTHENTICATION_FAILED", "Authentication failed", nil
	case errors.Is(err, billingapi.ErrForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden", nil
	case errors.Is(err, billingapi.ErrNotFound):
		return 404, "RESOURCE_NOT_FOUND", "Resource was not found", nil
	case errors.Is(err, billingapi.ErrTransitionInvalid):
		return 422, "BILLING_TRANSITION_INVALID", "Subscription transition is invalid", nil
	case errors.Is(err, billingapi.ErrImmutableConflict):
		return 409, "BILLING_IMMUTABLE_CONFLICT", "Immutable billing resource conflicts", nil
	case errors.Is(err, billingapi.ErrRevisionStale):
		return 409, "BILLING_REVISION_STALE", "Billing resource revision is stale", nil
	case errors.Is(err, billingapi.ErrIdempotencyReused):
		return 409, "IDEMPOTENCY_KEY_REUSED", "Idempotency key was reused", nil
	case errors.Is(err, billingapi.ErrIdempotencyProgress):
		return 409, "IDEMPOTENCY_IN_PROGRESS", "Operation is in progress", nil
	case errors.Is(err, billingapi.ErrConflict):
		return 409, "BILLING_CONFLICT", "Billing operation conflicts", nil
	case errors.Is(err, billingapi.ErrAuthorityUnavailable):
		return 503, "BILLING_AUTHORITY_UNAVAILABLE", "Billing authority is unavailable", nil
	case errors.Is(err, billingapi.ErrInvalid), errors.Is(err, billingRequestInvalid),
		errors.Is(err, httpinput.ErrInvalid):
		return 400, "REQUEST_INVALID", "Request is invalid", nil
	default:
		return 503, "BILLING_UNAVAILABLE", "Billing is unavailable", nil
	}
}

func correlationID(r *http.Request) requestmeta.CorrelationID {
	if r != nil {
		if value, found := requestmeta.CorrelationIDFromContext(r.Context()); found {
			return value
		}
	}
	return requestmeta.EntropyFailureCorrelationID
}
