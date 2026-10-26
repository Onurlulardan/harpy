package iam

import (
	executionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/restquery/execution"
	organizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/restquery/organization"
	policypostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/restquery/policy"
	rolepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/restquery/role"
	userpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/restquery/user"
	executionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/execution"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/organization"
	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/policy"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/role"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/user"
)

func buildIAMRESTQueryCatalogs() (
	*userapp.Service,
	*roleapp.Service,
	*organizationapp.Service,
	error,
) {
	users, err := userpostgres.New()
	if err != nil {
		return nil, nil, nil, err
	}
	userService, err := userapp.NewService(users)
	if err != nil {
		return nil, nil, nil, err
	}
	roles, err := rolepostgres.New()
	if err != nil {
		return nil, nil, nil, err
	}
	roleService, err := roleapp.NewService(roles)
	if err != nil {
		return nil, nil, nil, err
	}
	organizations, err := organizationpostgres.New()
	if err != nil {
		return nil, nil, nil, err
	}
	organizationService, err := organizationapp.NewService(organizations)
	if err != nil {
		return nil, nil, nil, err
	}
	return userService, roleService, organizationService, nil
}

func buildIAMRESTQueryPolicy() (*policyapp.Service, error) {
	return policypostgres.NewPolicyService()
}

func buildIAMRESTQueryExecution() (*executionapp.Executor, error) {
	driver, err := executionpostgres.New()
	if err != nil {
		return nil, err
	}
	return executionapp.NewExecutor(driver)
}
