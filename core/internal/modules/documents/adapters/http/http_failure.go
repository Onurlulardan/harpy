package documentshttp

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type failureDefinition struct {
	status        int
	code, message string
}

func (handler *Handler) writeFailure(w http.ResponseWriter, r *http.Request, err error) {
	definition := failureFor(err)
	if definition.status >= 500 {
		handler.report(errors.New("Documents HTTP operation failed"))
	}
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(definition.code,
		definition.message, handler.correlation(r), nil)
	if envelopeErr != nil {
		handler.report(errors.New("Documents HTTP error envelope failed"))
		return
	}
	handler.writeJSON(w, r, definition.status, envelope)
}

func failureFor(err error) failureDefinition {
	switch {
	case errors.Is(err, errBodyTooLarge):
		return failure(413, "REQUEST_BODY_TOO_LARGE", "Request body is too large")
	case errors.Is(err, errMediaUnsupported):
		return failure(415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported")
	case errors.Is(err, ErrAuthentication):
		return failure(401, "AUTHENTICATION_FAILED", "Authentication failed")
	case errors.Is(err, ErrForbidden):
		return failure(403, "AUTHORIZATION_DENIED", "Operation is forbidden")
	case errors.Is(err, ErrNotFound):
		return failure(404, "DOCUMENT_NOT_FOUND", "Document resource was not found")
	case errors.Is(err, ErrConflict):
		return failure(409, "DOCUMENT_STATE_CONFLICT", "Document state conflicts")
	case errors.Is(err, ErrClientUpgradeRequired):
		return failure(426, "CLIENT_UPGRADE_REQUIRED", "Asynchronous Documents API is required")
	case errors.Is(err, ErrContentRejected):
		return failure(422, "DOCUMENT_CONTENT_REJECTED", "Document content was rejected")
	case errors.Is(err, ErrScanUnavailable):
		return failure(503, "DOCUMENT_SCAN_UNAVAILABLE", "Document scanner is unavailable")
	case errors.Is(err, ErrInvalid):
		return failure(422, "DOCUMENT_INPUT_INVALID", "Document input is invalid")
	case errors.Is(err, errRequestInvalid), errors.Is(err, errDocumentInvalid):
		return failure(400, "REQUEST_INVALID", "Request is invalid")
	default:
		return failure(503, "DOCUMENTS_UNAVAILABLE", "Documents is unavailable")
	}
}

func failure(status int, code, message string) failureDefinition {
	return failureDefinition{status: status, code: code, message: message}
}
