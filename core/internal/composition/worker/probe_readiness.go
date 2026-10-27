package worker

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func probeReadiness(
	pool *postgres.Pool,
	policyReady func(context.Context) bool,
) health.Probe {
	return func(ctx context.Context) health.States {
		observation := pool.ProbeReadiness(ctx)
		policy := policyReady != nil && observation.DatabaseReady() &&
			policyReady(ctx)
		return health.States{
			Database:         observation.DatabaseReady(),
			DatabaseRole:     observation.DatabaseRoleReady(),
			DeploymentPolicy: policy,
			SchemaContract:   observation.SchemaContractReady(),
		}
	}
}

func allReady(states health.States) bool {
	return states.Database && states.DatabaseRole && states.DeploymentPolicy &&
		states.SchemaContract
}
