package tenancy

import (
	"context"

	policyadapter "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/deploymentpolicy"
	targetadapter "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/primarytarget"
	tenancyapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/deploymentpolicy"
	targetapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/primarytarget"
)

type DeploymentPolicy = tenancyapi.DeploymentPolicy
type PrimaryTenantTarget = tenancyapi.PrimaryTenantTarget
type DeploymentPolicyQuery = tenancyapi.DeploymentPolicyQuery
type PrimaryTenantTargetQuery = tenancyapi.PrimaryTenantTargetQuery

type DeploymentPolicyRow interface {
	Scan(...any) error
}

type DeploymentPolicyQueryRow func(
	context.Context,
	string,
	...any,
) DeploymentPolicyRow

func NewDeploymentPolicyQuery(
	queryRow DeploymentPolicyQueryRow,
) (DeploymentPolicyQuery, error) {
	if queryRow == nil {
		return nil, ErrDeploymentPolicyDependencies
	}
	reader, err := policyadapter.NewPolicyReader(policyadapter.QueryRow(func(
		ctx context.Context,
		statement string,
		arguments ...any,
	) policyadapter.Row {
		return queryRow(ctx, statement, arguments...)
	}))
	if err != nil {
		return nil, err
	}
	return policyapp.New(reader)
}

func NewPrimaryTenantTargetQuery(
	queryRow DeploymentPolicyQueryRow,
) (PrimaryTenantTargetQuery, error) {
	if queryRow == nil {
		return nil, ErrDeploymentPolicyDependencies
	}
	reader, err := targetadapter.NewReader(targetadapter.QueryRow(func(
		ctx context.Context,
		statement string,
		arguments ...any,
	) targetadapter.Row {
		return queryRow(ctx, statement, arguments...)
	}))
	if err != nil {
		return nil, err
	}
	return targetapp.New(reader)
}
