package iam

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type systemTransferFailure struct {
	status        int
	code, message string
}

func (handler *SystemTransferHTTP) writeTransferFailure(w http.ResponseWriter,
	r *http.Request, err error) {
	failure := mapSystemTransferFailure(err)
	if failure.status >= 500 {
		handler.report(errors.New("IAM transfer HTTP operation failed"))
	}
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(failure.code,
		failure.message, systemTransferCorrelationID(r).String(), nil)
	if envelopeErr == nil {
		envelopeErr = httpwire.WriteJSON(w, failure.status,
			systemTransferCorrelationID(r), envelope, false)
	}
	if envelopeErr != nil {
		handler.report(errors.New("IAM transfer HTTP failure write failed"))
	}
}

func mapSystemTransferFailure(err error) systemTransferFailure {
	switch {
	case errors.Is(err, ErrSystemTransferTooLarge), errors.Is(err, httpinput.ErrBodyTooLarge):
		return transferFailure(413, "IAM_TRANSFER_TOO_LARGE", "Transfer payload is too large")
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return transferFailure(415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported")
	case errors.Is(err, ErrSystemTransferAuthentication):
		return transferFailure(401, "AUTHENTICATION_FAILED", "Authentication failed")
	case errors.Is(err, ErrSystemTransferForbidden):
		return transferFailure(403, "AUTHORIZATION_DENIED", "Operation is forbidden")
	case errors.Is(err, ErrSystemTransferNotFound):
		return transferFailure(404, "IAM_TRANSFER_NOT_FOUND", "Transfer was not found")
	case errors.Is(err, ErrSystemTransferExpired):
		return transferFailure(410, "IAM_EXPORT_EXPIRED", "Export artifact is expired")
	case errors.Is(err, ErrSystemTransferConflict):
		return transferFailure(409, "IAM_TRANSFER_CONFLICT", "Transfer conflicts")
	case errors.Is(err, ErrSystemTransferBusy):
		return transferFailure(429, "IAM_TRANSFER_CAPACITY", "Transfer capacity is exhausted")
	case errors.Is(err, ErrSystemTransferInvalid):
		return transferFailure(422, "IAM_TRANSFER_INVALID", "Transfer input is invalid")
	case errors.Is(err, errSystemTransferRequest), errors.Is(err, httpinput.ErrInvalid):
		return transferFailure(400, "REQUEST_INVALID", "Request is invalid")
	default:
		return transferFailure(503, "IAM_TRANSFER_UNAVAILABLE", "Transfer is unavailable")
	}
}

func transferFailure(status int, code, message string) systemTransferFailure {
	return systemTransferFailure{status: status, code: code, message: message}
}
