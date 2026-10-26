package licenseauthority

import (
	"context"
	"fmt"
	"time"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

func (storage *store) AcquireRefreshTarget(
	ctx context.Context,
) (licenseapp.RefreshTarget, error) {
	if !storage.usable(ctx) {
		return licenseapp.RefreshTarget{}, licenseapp.ErrAuthorityState
	}
	var target licenseapp.RefreshTarget
	var delivery string
	err := storage.database.QueryRow(ctx, acquireRefreshTargetSQL).Scan(
		&target.Found, &target.Ready, &target.OperationID,
		&target.Binding.LicenseID, &target.Binding.LicenseeOrganizationID,
		&target.Binding.EnvironmentID, &delivery, &target.Revision,
		&target.RetryAt, &target.DatabaseNow,
	)
	if err != nil {
		return licenseapp.RefreshTarget{}, fmt.Errorf(
			"acquire billing license refresh target: %w", err,
		)
	}
	target.Binding.DeliveryKind = licenseapp.DeliveryKind(delivery)
	target.RetryAt = target.RetryAt.In(time.UTC)
	target.DatabaseNow = target.DatabaseNow.In(time.UTC)
	return target, nil
}
