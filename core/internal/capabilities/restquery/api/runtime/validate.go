package runtime

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
)

func Validate(request api.QueryRequest, catalogs api.CatalogSet) error {
	if request.Validate() != nil || catalogs.Validate() != nil {
		return fmt.Errorf("REST Query validation input is invalid")
	}
	_, err := resolver.Resolve(request, catalogs)
	return err
}
