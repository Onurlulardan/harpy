package billing

import (
	"context"

	licenseauthorityapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

type LocalLicenseDeliveryKind = licenseauthorityapp.DeliveryKind
type LocalRestoredLicenseLookup = licenseauthorityapp.RestoredLicenseLookup
type LocalRestoredLicenseState = licenseauthorityapp.RestoredLicenseState
type LocalRestoredLicenseReceiptStore = licenseauthorityapp.RestoredLicenseReceiptStore

var (
	ErrLocalLicenseRefreshRetryable = licenseauthorityapp.ErrRefreshRetryable
	ErrLocalLicenseRefreshFenced    = licenseauthorityapp.ErrRefreshFenced
)

func ParseUUID(value string) (UUID, error) {
	return uuid.Parse(value)
}

func NewUUID() UUID {
	return uuid.New()
}

func NewUUIDv7() (UUID, error) {
	return uuid.NewV7()
}

func ReadLocalRestoredLicenseReceipt(
	ctx context.Context,
	store LocalRestoredLicenseReceiptStore,
	lookup LocalRestoredLicenseLookup,
) (LocalRestoredLicenseState, error) {
	return licenseauthorityapp.GetRestoredLicenseReceipt(ctx, store, lookup)
}
