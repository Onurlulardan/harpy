package platformauthority

import (
	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type OnPremisePlannerDependencies struct {
	Pool                *postgresadapter.Pool
	Clock               ports.Clock
	RecoveryTrustBundle []byte
	RecoveryControlSeed []byte
}
