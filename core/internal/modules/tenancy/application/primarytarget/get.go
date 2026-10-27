package primarytarget

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (query *Queries) GetPrimaryTenantTarget(
	ctx context.Context,
) (tenantapi.PrimaryTenantTarget, error) {
	if query == nil || query.reader == nil || ctx == nil || ctx.Err() != nil {
		return tenantapi.PrimaryTenantTarget{},
			tenantapi.ErrPrimaryTenantTargetUnavailable
	}
	target, found, err := query.reader.GetPrimaryTenantTarget(ctx)
	if err != nil {
		return tenantapi.PrimaryTenantTarget{}, fmt.Errorf(
			"%w: read persisted primary tenant target: %v",
			tenantapi.ErrPrimaryTenantTargetUnavailable,
			err,
		)
	}
	if !found {
		return tenantapi.PrimaryTenantTarget{},
			tenantapi.ErrPrimaryTenantTargetMissing
	}
	if !target.IsValid() {
		return tenantapi.PrimaryTenantTarget{},
			tenantapi.ErrPrimaryTenantTargetUnavailable
	}
	return target, nil
}

var _ tenantapi.PrimaryTenantTargetQuery = (*Queries)(nil)
