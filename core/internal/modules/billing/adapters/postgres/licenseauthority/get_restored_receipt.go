package licenseauthority

import (
	"context"
	"fmt"
	"time"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

func (storage *store) GetRestoredLicenseReceipt(
	ctx context.Context,
	operationID uuid.UUID,
	environmentID uuid.UUID,
) (licenseapp.RestoredLicenseState, error) {
	if !storage.usable(ctx) || operationID == uuid.Nil || environmentID == uuid.Nil {
		return licenseapp.RestoredLicenseState{}, licenseapp.ErrAuthorityState
	}
	var result licenseapp.RestoredLicenseState
	err := storage.database.QueryRow(
		ctx, readRestoredLicenseReceiptSQL, operationID, environmentID,
	).Scan(
		&result.Found, &result.OperationID, &result.RestoreReceiptID,
		&result.RestoreReceiptSHA256, &result.LicenseBindingID,
		&result.LicenseID, &result.LicenseeOrganizationID,
		&result.EnvironmentID, &result.PreviousSequence,
		&result.PreviousSHA256, &result.AcceptedSequence,
		&result.AcceptedSHA256, &result.AuthorityRevision,
		&result.DatabaseNow,
	)
	if err != nil {
		return licenseapp.RestoredLicenseState{}, fmt.Errorf(
			"read restored license receipt: %w", err,
		)
	}
	result.DatabaseNow = result.DatabaseNow.In(time.UTC)
	return result, nil
}
