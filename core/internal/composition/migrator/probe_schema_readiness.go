package migrator

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func probeSchemaReadiness(pool *postgres.Pool) health.Probe {
	return func(ctx context.Context) health.States {
		observation := pool.ProbeReadiness(ctx)
		return health.States{
			Database:         observation.DatabaseReady(),
			DatabaseRole:     observation.DatabaseRoleReady(),
			DeploymentPolicy: true,
			SchemaContract:   observation.SchemaContractReady(),
		}
	}
}

func schemaReady(states health.States) bool {
	return states.Database && states.DatabaseRole && states.DeploymentPolicy &&
		states.SchemaContract
}
