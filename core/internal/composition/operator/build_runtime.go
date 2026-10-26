package operator

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func BuildRuntime(pool *platformpostgres.Pool) (*tenancy.OperatorRuntime, error) {
	return tenancy.BindOperatorRuntime(pool)
}
