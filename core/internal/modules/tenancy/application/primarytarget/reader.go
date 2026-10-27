package primarytarget

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type Reader interface {
	GetPrimaryTenantTarget(
		context.Context,
	) (tenantapi.PrimaryTenantTarget, bool, error)
}
