package api

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func probeReadiness(
	pool *postgres.Pool,
	policyReady func(context.Context) bool,
	localChecks ...func() bool,
) health.Probe {
	return func(ctx context.Context) health.States {
		if policyReady == nil || !localDependenciesReady(localChecks) {
			return health.States{}
		}
		observation := pool.ProbeReadiness(ctx)
		policy := observation.DatabaseReady() && policyReady(ctx)
		return health.States{
			Database:         observation.DatabaseReady(),
			DatabaseRole:     observation.DatabaseRoleReady(),
			DeploymentPolicy: policy,
		}
	}
}

func localDependenciesReady(checks []func() bool) bool {
	for _, check := range checks {
		if check == nil || !check() {
			return false
		}
	}
	return true
}

func allReady(states health.States) bool {
	return states.Database && states.DatabaseRole && states.DeploymentPolicy
}

func readinessFailure(states health.States) string {
	switch {
	case !states.Database:
		return "database readiness probe failed"
	case !states.DatabaseRole:
		return "database role readiness probe failed"
	case !states.DeploymentPolicy:
		return "deployment policy readiness probe failed"
	default:
		return "initial readiness probe failed"
	}
}
