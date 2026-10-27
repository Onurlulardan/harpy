package licenseauthority

import (
	"context"
	"fmt"
	"time"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

func (storage *store) ResolveRestoreBinding(
	ctx context.Context,
	environmentID uuid.UUID,
) (licenseapp.RestoreBinding, error) {
	if !storage.usable(ctx) || environmentID == uuid.Nil {
		return licenseapp.RestoreBinding{}, licenseapp.ErrAuthorityState
	}
	var result licenseapp.RestoreBinding
	var delivery string
	err := storage.database.QueryRow(ctx, readRestoreBindingSQL, environmentID).Scan(
		&result.Binding.LicenseID, &result.Binding.LicenseeOrganizationID,
		&result.Binding.EnvironmentID, &delivery, &result.Issuer,
		&result.Sequence, &result.SHA256, &result.AuthorityRevision,
		&result.DatabaseNow,
	)
	if err != nil {
		return licenseapp.RestoreBinding{}, fmt.Errorf(
			"read billing restored license binding: %w", err,
		)
	}
	result.Binding.DeliveryKind = licenseapp.DeliveryKind(delivery)
	result.DatabaseNow = result.DatabaseNow.In(time.UTC)
	return result, nil
}
