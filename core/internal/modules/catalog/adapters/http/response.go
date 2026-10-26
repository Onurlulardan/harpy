package cataloghttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

func (runtime *Runtime) Success(writer http.ResponseWriter, request *http.Request,
	status int, value any,
) {
	body, err := json.Marshal(struct {
		Success bool `json:"success"`
		Data    any  `json:"data"`
	}{Success: true, Data: value})
	if err != nil || writeJSON(writer, status, runtime.correlation(request), body) != nil {
		runtime.report(errors.New("catalog HTTP response write failed"))
	}
}

func (runtime *Runtime) Failure(writer http.ResponseWriter, request *http.Request, err error) {
	status, code, message := failureDefinition(err)
	if status >= 500 {
		runtime.report(errors.New("catalog HTTP operation failed"))
	}
	correlation := runtime.correlation(request)
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(code, message, correlation, nil)
	if envelopeErr == nil {
		var body []byte
		body, envelopeErr = json.Marshal(envelope)
		if envelopeErr == nil {
			envelopeErr = writeJSON(writer, status, correlation, body)
		}
	}
	if envelopeErr != nil {
		runtime.report(errors.New("catalog HTTP failure write failed"))
	}
}

func failureDefinition(err error) (int, string, string) {
	switch {
	case errors.Is(err, ErrAuthentication):
		return 401, "AUTHENTICATION_FAILED", "Authentication failed"
	case errors.Is(err, ErrForbidden), errors.Is(err, lifecycleapp.ErrForbidden),
		errors.Is(err, menuapp.ErrForbidden), errors.Is(err, packageregistry.ErrForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden"
	case errors.Is(err, ErrNotFound), errors.Is(err, lifecycleapp.ErrNotFound),
		errors.Is(err, menuapp.ErrNotFound), errors.Is(err, packageregistry.ErrNotFound):
		return 404, "CATALOG_NOT_FOUND", "Catalog resource was not found"
	case errors.Is(err, lifecycleapp.ErrOwnerNotV2):
		return 409, "CATALOG_OWNER_NOT_V2", "Component owner is not ready"
	case errors.Is(err, lifecycleapp.ErrDependencyBlocked):
		return 409, "CATALOG_DEPENDENCY_BLOCKED", "Catalog dependency blocks the operation"
	case errors.Is(err, lifecycleapp.ErrBackupBlocked):
		return 409, "CATALOG_BACKUP_BLOCKED", "Verified backup is required"
	case errors.Is(err, lifecycleapp.ErrOwnerReceiptStale):
		return 409, "CATALOG_OWNER_RECEIPT_STALE", "Component owner receipt is stale"
	case errors.Is(err, lifecycleapp.ErrDemoPlanStale):
		return 409, "CATALOG_DEMO_PLAN_STALE", "Catalog demo removal plan is stale"
	case errors.Is(err, ErrConflict), errors.Is(err, lifecycleapp.ErrConflict), errors.Is(err, menuapp.ErrConflict):
		return 409, "CATALOG_STATE_CONFLICT", "Catalog state conflicts"
	case errors.Is(err, ErrInvalid), errors.Is(err, lifecycleapp.ErrInvalid),
		errors.Is(err, menuapp.ErrInvalid), errors.Is(err, packageregistry.ErrInvalid):
		return 400, "REQUEST_INVALID", "Request is invalid"
	default:
		return 503, "CATALOG_UNAVAILABLE", "Catalog is unavailable"
	}
}

func writeJSON(writer http.ResponseWriter, status int, correlation string, body []byte) error {
	if !json.Valid(body) {
		return fmt.Errorf("invalid catalog HTTP response")
	}
	header := writer.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Correlation-ID", correlation)
	header.Set("Content-Length", strconv.Itoa(len(body)))
	writer.WriteHeader(status)
	_, err := writer.Write(body)
	return err
}
