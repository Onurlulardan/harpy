package billing

import (
	"errors"
	"net/http"
	"strings"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/google/uuid"
)

const maximumCommandBodyBytes int64 = 64 * 1024

var billingRequestInvalid = errors.New("billing HTTP request is invalid")

func decodeCommand[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var input T
	if err := httpinput.DecodeJSON(w, r, maximumCommandBodyBytes, &input); err != nil {
		return input, err
	}
	return input, nil
}

func pathUUID(r *http.Request, name string) (string, bool) {
	if r == nil || name == "" {
		return "", false
	}
	value := r.PathValue(name)
	parsed, err := uuid.Parse(value)
	return value, err == nil && parsed != uuid.Nil && parsed.String() == value
}

func idempotencyKey(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	values := r.Header.Values("Idempotency-Key")
	if len(values) != 1 || len(values[0]) < 1 || len(values[0]) > 128 {
		return "", false
	}
	for _, character := range values[0] {
		if character < 0x21 || character > 0x7e {
			return "", false
		}
	}
	return values[0], true
}

func requiredText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && strings.TrimSpace(value) == value
}

func requireAccess(
	handler *Handler,
	r *http.Request,
	permission billingapi.Permission,
	operation string,
) (billingapi.Access, error) {
	access, err := handler.accessFor(r, permission, operation)
	if err != nil {
		return billingapi.Access{}, err
	}
	return access, nil
}
