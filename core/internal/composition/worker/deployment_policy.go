package worker

import (
	"context"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func bindDeploymentPolicy(
	pool *platformpostgres.Pool,
	expected config.DeploymentMode,
) (*tenancymodule.DeploymentPolicyRuntime, error) {
	policy, err := platformpostgres.BindTenancyPolicyReader(
		pool,
		func(query platformpostgres.TenancyPolicyQueryRow) (
			tenancymodule.DeploymentPolicyQuery,
			error,
		) {
			return tenancymodule.NewDeploymentPolicyQuery(workerPolicyRow(query))
		},
	)
	if err != nil {
		return nil, err
	}
	target, err := platformpostgres.BindTenancyPolicyReader(
		pool,
		func(query platformpostgres.TenancyPolicyQueryRow) (
			tenancymodule.PrimaryTenantTargetQuery,
			error,
		) {
			return tenancymodule.NewPrimaryTenantTargetQuery(workerPolicyRow(query))
		},
	)
	if err != nil {
		return nil, err
	}
	return tenancymodule.NewDeploymentPolicyRuntime(string(expected), policy, target)
}

func workerPolicyRow(
	query platformpostgres.TenancyPolicyQueryRow,
) tenancymodule.DeploymentPolicyQueryRow {
	return func(
		ctx context.Context,
		statement string,
		arguments ...any,
	) tenancymodule.DeploymentPolicyRow {
		return query.QueryRow(ctx, statement, arguments...)
	}
}
