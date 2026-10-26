package catalog

import (
	"time"

	catalogownerpostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/catalogowner"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	"github.com/google/uuid"
)

func NewCatalogComponentOwner() (*catalogownerpostgres.Owner, error) {
	values := catalogComponentValues{}
	return catalogownerpostgres.New(values, values)
}

type catalogComponentValues struct{}

func (catalogComponentValues) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func (catalogComponentValues) NewUUIDv7() (uuid.UUID, error) {
	return randomid.NewUUIDv7(time.Now().UTC())
}

var _ componentowner.Provisioner = (*catalogownerpostgres.Owner)(nil)
