package billing

import (
	"strings"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func subscriptionMatches(
	provided string,
	current billingapi.TenantSubscriptionSnapshot,
) error {
	if provided == "" {
		return nil
	}
	identifier, valid := parseV7(provided)
	if !valid {
		return billingapi.ErrInvalid
	}
	if identifier != current.ID {
		return billingapi.ErrNotFound
	}
	return nil
}

func validReason(value string) bool {
	return value != "" && len(value) <= 512 && strings.TrimSpace(value) == value
}
