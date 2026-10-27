package primarytarget

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (targetReader *reader) GetPrimaryTenantTarget(
	ctx context.Context,
) (tenantapi.PrimaryTenantTarget, bool, error) {
	if !targetReader.usable(ctx) {
		return tenantapi.PrimaryTenantTarget{}, false, errTargetUnavailable
	}
	var environmentID *string
	var mode *string
	var tenantID *string
	var purpose *string
	var slug *string
	var found bool
	err := targetReader.queryRow(ctx, getPrimaryTenantTargetSQL).Scan(
		&environmentID,
		&mode,
		&tenantID,
		&purpose,
		&slug,
		&found,
	)
	if err != nil {
		return tenantapi.PrimaryTenantTarget{}, false,
			fmt.Errorf("read tenancy primary tenant target: %w", err)
	}
	if !found {
		return tenantapi.PrimaryTenantTarget{}, false, nil
	}
	if environmentID == nil || mode == nil || tenantID == nil ||
		purpose == nil || slug == nil {
		return tenantapi.PrimaryTenantTarget{}, false, errTargetUnavailable
	}
	target, err := tenantapi.NewPrimaryTenantTarget(
		*environmentID,
		tenantapi.DeploymentMode(*mode),
		*tenantID,
		tenantapi.PrimaryTenantPurpose(*purpose),
		*slug,
	)
	if err != nil {
		return tenantapi.PrimaryTenantTarget{}, false, errTargetUnavailable
	}
	return target, true, nil
}
