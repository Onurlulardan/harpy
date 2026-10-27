package api

import (
	"errors"
	"net/http"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

var (
	errSystemResourceRecordHTTPRequestInvalid = errors.New(
		"system resource record HTTP request is invalid",
	)
	errSystemResourceRecordHTTPPrecondition = errors.New(
		"system resource record HTTP precondition is required",
	)
)

type systemResourceRecordHTTPFailureDefinition struct {
	status  int
	code    string
	message string
}

type systemResourceRecordHTTPErrorEnvelope struct {
	Success bool                               `json:"success"`
	Error   systemResourceRecordHTTPErrorEntry `json:"error"`
}

type systemResourceRecordHTTPErrorEntry struct {
	Code          string                          `json:"code"`
	Message       string                          `json:"message"`
	Issues        []systemResourceRecordHTTPIssue `json:"issues"`
	CorrelationID string                          `json:"correlation_id"`
}

type systemResourceRecordHTTPIssue struct {
	Path string `json:"path"`
	Code string `json:"code"`
}

func (handler *systemResourceRecordHTTPHandler) writeRecordHTTPFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	var queryFailure iammodule.SystemResourceQueryFailure
	if errors.As(err, &queryFailure) {
		handler.writeRecordHTTPQueryFailure(w, r, queryFailure)
		return
	}
	definition := systemResourceRecordHTTPFailure(err)
	if definition.status >= 500 {
		handler.report(errors.New("system resource record HTTP operation failed"))
	}
	envelope := newSystemResourceRecordHTTPErrorEnvelope(
		definition.code, definition.message, r, nil,
	)
	envelopeErr := httpwire.WriteJSON(
		w, definition.status, layoutHTTPCorrelationID(r), envelope, false,
	)
	if envelopeErr != nil {
		handler.report(errors.New("system resource record HTTP failure write failed"))
	}
}

func (handler *systemResourceRecordHTTPHandler) writeRecordHTTPQueryFailure(
	w http.ResponseWriter, r *http.Request, failure iammodule.SystemResourceQueryFailure,
) {
	issues := make([]systemResourceRecordHTTPIssue, 0, len(failure.Issues()))
	for _, issue := range failure.Issues() {
		issues = append(issues, systemResourceRecordHTTPIssue{
			Path: issue.Path, Code: issue.Code,
		})
	}
	envelope := newSystemResourceRecordHTTPErrorEnvelope(
		failure.Code(), "Query is invalid", r, issues,
	)
	err := httpwire.WriteJSON(
		w, failure.HTTPStatus(), layoutHTTPCorrelationID(r), envelope, false,
	)
	if err != nil {
		handler.report(errors.New("system resource record query failure write failed"))
	}
}

func newSystemResourceRecordHTTPErrorEnvelope(
	code string,
	message string,
	r *http.Request,
	issues []systemResourceRecordHTTPIssue,
) systemResourceRecordHTTPErrorEnvelope {
	if issues == nil {
		issues = []systemResourceRecordHTTPIssue{}
	}
	return systemResourceRecordHTTPErrorEnvelope{
		Error: systemResourceRecordHTTPErrorEntry{
			Code: code, Message: message, Issues: issues,
			CorrelationID: layoutHTTPCorrelationID(r).String(),
		},
	}
}

func systemResourceRecordHTTPFailure(
	err error,
) systemResourceRecordHTTPFailureDefinition {
	switch {
	case errors.Is(err, httpinput.ErrBodyTooLarge):
		return recordHTTPFailure(413, "REQUEST_BODY_TOO_LARGE", "Request body is too large")
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return recordHTTPFailure(415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported")
	case errors.Is(err, systemresources.ErrLayoutAccessAuthentication):
		return recordHTTPFailure(401, "AUTHENTICATION_FAILED", "Authentication failed")
	case errors.Is(err, systemresources.ErrLayoutAccessForbidden),
		errors.Is(err, systemresources.ErrForbidden):
		return recordHTTPFailure(403, "AUTHORIZATION_DENIED", "Operation is forbidden")
	case errors.Is(err, systemresources.ErrNotFound):
		return recordHTTPFailure(404, "RESOURCE_NOT_FOUND", "Resource was not found")
	case errors.Is(err, systemresources.ErrConflict):
		return recordHTTPFailure(409, "RESOURCE_CONFLICT", "Resource operation conflicts")
	case errors.Is(err, errSystemResourceRecordHTTPPrecondition):
		return recordHTTPFailure(428, "PRECONDITION_REQUIRED", "Revision precondition is required")
	case errors.Is(err, systemresources.ErrContract):
		return recordHTTPFailure(422, "SYSTEM_RESOURCE_INPUT_INVALID", "System resource input is invalid")
	case errors.Is(err, errSystemResourceRecordHTTPRequestInvalid),
		errors.Is(err, httpinput.ErrInvalid):
		return recordHTTPFailure(400, "REQUEST_INVALID", "Request is invalid")
	case errors.Is(err, systemresources.ErrLayoutAccessUnavailable):
		return recordHTTPFailure(503, "IAM_AUTHORITY_UNAVAILABLE", "IAM authority is unavailable")
	default:
		return recordHTTPFailure(503, "IAM_UNAVAILABLE", "IAM is unavailable")
	}
}

func recordHTTPFailure(
	status int, code, message string,
) systemResourceRecordHTTPFailureDefinition {
	return systemResourceRecordHTTPFailureDefinition{
		status: status, code: code, message: message,
	}
}
