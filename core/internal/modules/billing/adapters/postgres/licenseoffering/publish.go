package licenseoffering

import (
	"context"
	"fmt"

	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseoffering"
)

func (storage *store) PublishLicenseOffering(
	ctx context.Context,
	request offeringapp.PublishRequest,
) (offeringapp.Receipt, error) {
	if !storage.usable(ctx) {
		return offeringapp.Receipt{}, offeringapp.ErrConfiguration
	}
	binding := request.Binding
	receipt := offeringapp.Receipt{}
	err := storage.database.QueryRow(ctx, publishLicenseOfferingSQL,
		binding.OperationID, binding.EnvironmentID, binding.LicenseID,
		binding.LicenseeOrganizationID, binding.DeploymentMode,
		binding.DeliveryKind, binding.Issuer, binding.Sequence,
		binding.SnapshotSHA256, binding.OfferingVersionID,
		binding.Revision, binding.OfferingSHA256, binding.EffectiveAt,
		binding.ManifestSHA256, request.Offering.Bytes,
	).Scan(
		&receipt.OperationID, &receipt.OfferingVersionID,
		&receipt.Revision, &receipt.OfferingSHA256, &receipt.Replay,
	)
	if err != nil {
		return offeringapp.Receipt{}, fmt.Errorf(
			"persist license-derived offering: %w", err,
		)
	}
	return receipt, nil
}
