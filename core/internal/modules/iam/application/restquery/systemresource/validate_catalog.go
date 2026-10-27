package systemresourcequery

import (
	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

func ValidateCatalogRequest(published catalog.Published, request Request) error {
	value, requestErr := request.Value()
	set, catalogErr := published.CatalogSet()
	if requestErr != nil || catalogErr != nil {
		return ErrContract
	}
	if err := queryruntime.Validate(value, set); err != nil {
		return NormalizeError(err)
	}
	return nil
}
