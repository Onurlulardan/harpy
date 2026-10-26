package audithttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
)

type successEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type failureEnvelope struct {
	Success bool        `json:"success"`
	Error   failureBody `json:"error"`
}

type failureBody struct{ Code, Message, CorrelationID string }

func (body failureBody) MarshalJSON() ([]byte, error) {
	type wire struct {
		Code          string `json:"code"`
		Message       string `json:"message"`
		CorrelationID string `json:"correlation_id"`
	}
	return json.Marshal(wire{Code: body.Code, Message: body.Message,
		CorrelationID: body.CorrelationID})
}

func (handler *Handler) writeSuccess(
	response http.ResponseWriter, request *http.Request, data any,
) {
	handler.writeJSON(response, request, http.StatusOK,
		successEnvelope{Success: true, Data: data})
}

func (handler *Handler) writeFailure(
	response http.ResponseWriter, request *http.Request, err error,
) {
	status, code := failureStatus(err)
	handler.writeJSON(response, request, status, failureEnvelope{Error: failureBody{
		Code: code, Message: "Request could not be completed",
		CorrelationID: handler.dependencies.Correlation(request)}})
}

func failureStatus(err error) (int, string) {
	switch {
	case errors.Is(err, application.ErrAuthentication):
		return http.StatusUnauthorized, "AUTHENTICATION_REQUIRED"
	case errors.Is(err, application.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, application.ErrInvalid):
		return http.StatusUnprocessableEntity, "INPUT_INVALID"
	case errors.Is(err, application.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, application.ErrConflict):
		return http.StatusConflict, "STATE_CONFLICT"
	default:
		return http.StatusServiceUnavailable, "AUTHORITY_UNAVAILABLE"
	}
}

func (handler *Handler) writeJSON(
	response http.ResponseWriter, request *http.Request, status int, payload any,
) {
	body, err := json.Marshal(payload)
	if err != nil {
		handler.dependencies.Report(errors.New("audit HTTP response marshal failed"))
		return
	}
	header := response.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Correlation-ID", handler.dependencies.Correlation(request))
	header.Set("Content-Length", strconv.Itoa(len(body)))
	response.WriteHeader(status)
	if _, err = response.Write(body); err != nil {
		handler.dependencies.Report(errors.New("audit HTTP response write failed"))
	}
}
