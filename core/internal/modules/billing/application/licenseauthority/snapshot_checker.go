package licenseauthority

import (
	"context"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/google/uuid"
)

type SnapshotChecker struct{ keys KeyResolver }

func NewSnapshotChecker(keys KeyResolver) (*SnapshotChecker, error) {
	if keys == nil {
		return nil, ErrUntrustedKey
	}
	return &SnapshotChecker{keys: keys}, nil
}

func (checker *SnapshotChecker) CheckLicenseSnapshot(
	ctx context.Context,
	document []byte,
	binding billingapi.LicenseVerificationBinding,
	databaseNow time.Time,
) (billingapi.VerifiedLicenseSnapshot, error) {
	licenseID, err := uuid.Parse(binding.LicenseID)
	if err != nil {
		return nil, ErrWrongBinding
	}
	organizationID, err := uuid.Parse(binding.LicenseeOrganizationID)
	if err != nil {
		return nil, ErrWrongBinding
	}
	environmentID, err := uuid.Parse(binding.EnvironmentID)
	if err != nil {
		return nil, ErrWrongBinding
	}
	snapshot, err := Verify(ctx, document, Binding{
		LicenseID: licenseID, LicenseeOrganizationID: organizationID,
		EnvironmentID: environmentID, DeliveryKind: DeliveryKind(binding.DeliveryKind),
	}, databaseNow, checker.keys)
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

var _ billingapi.LicenseSnapshotChecker = (*SnapshotChecker)(nil)
