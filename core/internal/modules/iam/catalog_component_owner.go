package iam

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	catalogownerpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/catalogowner"
)

func NewCatalogComponentOwner() (*catalogownerpostgres.Owner, error) {
	return catalogownerpostgres.New()
}

var _ componentowner.Provisioner = (*catalogownerpostgres.Owner)(nil)
