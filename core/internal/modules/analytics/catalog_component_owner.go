package analytics

import (
	cataloglifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/cataloglifecycle"
	catalogownerpostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/catalogowner"
	catalogownerapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/catalogcomponentowner"
	cataloglegacyapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/cataloglegacyowner"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
)

func NewCatalogComponentOwner() (*catalogownerapp.Owner, error) {
	base, err := catalogownerpostgres.New()
	if err != nil {
		return nil, err
	}
	legacy, err := cataloglegacyapp.New(base)
	if err != nil {
		return nil, err
	}
	return catalogownerapp.New(legacy, cataloglifecyclepostgres.NewBinder())
}

var _ componentowner.Provisioner = (*catalogownerapp.Owner)(nil)
