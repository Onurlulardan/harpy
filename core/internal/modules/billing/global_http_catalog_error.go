package billing

import (
	"errors"

	catalogquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func normalizeCatalogError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, catalogquery.ErrNotFound):
		return billingapi.ErrNotFound
	case errors.Is(err, catalogquery.ErrRequest):
		return billingapi.ErrInvalid
	default:
		return billingapi.ErrAuthorityUnavailable
	}
}
