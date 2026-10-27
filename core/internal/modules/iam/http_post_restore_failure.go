package iam

import (
	"errors"
	"fmt"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *PostRestoreHTTP) writePostRestoreInputFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	status, code, message := http.StatusBadRequest, "REQUEST_INVALID", "Request is invalid"
	if errors.Is(err, httpinput.ErrBodyTooLarge) {
		status, code, message = http.StatusRequestEntityTooLarge,
			"REQUEST_BODY_TOO_LARGE", "Request body is too large"
	} else if errors.Is(err, httpinput.ErrMediaTypeUnsupported) {
		status, code, message = http.StatusUnsupportedMediaType,
			"MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"
	}
	handler.writePostRestoreFailure(w, r, status, code, message)
}

func (handler *PostRestoreHTTP) writePostRestoreApplicationFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	handler.report(fmt.Errorf("IAM post-restore operation rejected: %w", err))
	handler.writePostRestoreFailure(w, r, http.StatusServiceUnavailable,
		"POST_RESTORE_AUTHENTICATION_UNAVAILABLE", "Post-restore authentication is unavailable")
}

func (handler *PostRestoreHTTP) writePostRestoreFailure(
	w http.ResponseWriter, r *http.Request, status int, code, message string,
) {
	correlationID := authHTTPCorrelationID(r)
	envelope, err := runtimev1.NewErrorEnvelope(code, message, correlationID.String(), nil)
	if err == nil {
		err = httpwire.WriteJSON(w, status, correlationID, envelope, false)
	}
	if err != nil {
		handler.report(fmt.Errorf("write IAM post-restore failure: %w", err))
	}
}
