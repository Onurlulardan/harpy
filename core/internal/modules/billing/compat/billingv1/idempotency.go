package billingv1

import (
	"fmt"
	"net/http"
)

func (handler *Handler) idempotencyKey(r *http.Request) (string, error) {
	values := r.Header.Values("Idempotency-Key")
	if len(values) == 1 {
		if validIdempotencyKey(values[0]) {
			return values[0], nil
		}
		return "", ErrRequestInvalid
	}
	if len(values) > 1 {
		return "", ErrRequestInvalid
	}
	identifier, err := handler.nextID()
	if err != nil || identifier.Version() != 7 {
		return "", fmt.Errorf("generate Billing V1 idempotency key: %w", err)
	}
	return "billing-v1-" + identifier.String(), nil
}

func validIdempotencyKey(value string) bool {
	if len(value) < 1 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if character < 0x21 || character > 0x7e {
			return false
		}
	}
	return true
}
