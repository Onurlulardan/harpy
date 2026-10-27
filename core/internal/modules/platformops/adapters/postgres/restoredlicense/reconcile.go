package restoredlicense

import (
	"context"
	"fmt"
	"time"

	restoredlicensev1 "github.com/Onurlulardan/harpy/contracts/restoredlicense/v1"
	grantapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/restoredlicense"
)

type databaseRow interface {
	Scan(...any) error
}

type store struct {
	queryRow func(context.Context, string, ...any) databaseRow
}

var _ grantapp.Store = (*store)(nil)

func (store *store) ReconcileRestoredLicense(
	ctx context.Context,
	input grantapp.Input,
) (grantapp.Result, error) {
	if store == nil || store.queryRow == nil || ctx == nil || ctx.Err() != nil {
		return grantapp.Result{}, fmt.Errorf("platformops restored license store is unavailable")
	}
	result := grantapp.Result{}
	binding := &result.Grant.Binding
	err := store.queryRow(
		ctx, mintRestoredLicenseGrantSQL, input.OperationID, input.EnvironmentID,
		input.RestoreReceiptID, input.RestoreReceiptSHA256,
		input.LicenseID, input.LicenseeOrganizationID,
		input.DeliveryKind, input.Issuer, input.PreviousSequence,
		input.PreviousSHA256, input.NextSequence, input.NextSHA256,
		restoredlicensev1.ManifestSHA256(), input.AuthorityRevision,
	).Scan(
		&binding.OperationID, &binding.EnvironmentID, &binding.RestoreReceiptID,
		&binding.RestoreReceiptSHA256, &binding.LicenseBindingID,
		&binding.LicenseID, &binding.LicenseeOrganizationID,
		&binding.DeliveryKind, &binding.Issuer, &binding.PreviousSequence,
		&binding.PreviousSHA256, &binding.NextSequence, &binding.NextSHA256,
		&binding.ManifestSHA256, &result.Grant.Revision,
		&result.Grant.LeaseUntil, &result.DatabaseNow,
	)
	if err != nil {
		return grantapp.Result{}, fmt.Errorf(
			"mint platformops restored license grant: %w", err,
		)
	}
	result.Grant.LeaseUntil = result.Grant.LeaseUntil.In(time.UTC)
	result.DatabaseNow = result.DatabaseNow.In(time.UTC)
	return result, nil
}
