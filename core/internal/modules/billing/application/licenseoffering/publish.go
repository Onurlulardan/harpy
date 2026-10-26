package licenseoffering

import (
	"context"
	"fmt"
	"time"

	licensev2 "github.com/Onurlulardan/harpy/contracts/license/v2"
	licensebootstrapv1 "github.com/Onurlulardan/harpy/contracts/licensebootstrap/v1"
	licensebootstrap "github.com/Onurlulardan/harpy/core/internal/capabilities/licensebootstrap/api"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
)

type Publisher struct {
	store Store
	check billingapi.LicenseSnapshotChecker
}

func NewPublisher(store Store, check billingapi.LicenseSnapshotChecker) (*Publisher, error) {
	if store == nil || check == nil {
		return nil, ErrConfiguration
	}
	return &Publisher{store: store, check: check}, nil
}

func (publisher *Publisher) Publish(
	ctx context.Context,
	scope licensebootstrap.Scope,
	signedDocument []byte,
	databaseNow time.Time,
) (Receipt, error) {
	binding := scope.Binding()
	licenseBinding := billingapi.LicenseVerificationBinding{
		LicenseID:              binding.LicenseID.String(),
		LicenseeOrganizationID: binding.LicenseeOrganizationID.String(),
		EnvironmentID:          binding.EnvironmentID.String(),
		DeliveryKind:           licensev2.DeliveryKind(binding.DeliveryKind),
	}
	snapshot, err := publisher.check.CheckLicenseSnapshot(
		ctx, signedDocument, licenseBinding, databaseNow,
	)
	if err != nil || !snapshotMatches(binding, snapshot) {
		return Receipt{}, ErrBinding
	}
	offering, err := BuildDocument(Plan{
		OfferingVersionID: binding.OfferingVersionID,
		Revision:          binding.Revision, EffectiveAt: binding.EffectiveAt,
	}, snapshot)
	if err != nil || offering.SHA256 != binding.OfferingSHA256 {
		return Receipt{}, ErrProjection
	}
	if err = scope.Consume(licensebootstrapv1.PublishLicenseDerivedOffering); err != nil {
		return Receipt{}, fmt.Errorf("consume license bootstrap command: %w", err)
	}
	receipt, err := publisher.store.PublishLicenseOffering(ctx, PublishRequest{
		Binding: binding, Offering: offering,
	})
	if err != nil {
		return Receipt{}, fmt.Errorf("publish license-derived offering: %w", err)
	}
	if !validReceipt(binding, receipt) {
		return Receipt{}, ErrReceipt
	}
	return receipt, nil
}
