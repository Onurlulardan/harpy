package policypostgres

import (
	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	iamauthorization "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/policy"
)

func NewPolicyService() (*policyapp.Service, error) {
	ports, err := newPorts()
	if err != nil {
		return nil, catalog.ErrContract
	}
	return newPolicyService(ports, ports)
}

func NewPolicyServiceForSnapshot(
	catalogs policyapp.CatalogPort,
	versions policyapp.VersionPort,
) (*policyapp.Service, error) {
	return newPolicyService(catalogs, versions)
}

func newPolicyService(
	catalogs policyapp.CatalogPort,
	versions policyapp.VersionPort,
) (*policyapp.Service, error) {
	evidence, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, catalog.ErrContract
	}
	authorizer, err := iamauthorization.NewAuthorizer(evidence)
	if err != nil {
		return nil, catalog.ErrContract
	}
	service, err := policyapp.NewService(authorizer, catalogs, versions)
	if err != nil {
		return nil, catalog.ErrContract
	}
	return service, nil
}
