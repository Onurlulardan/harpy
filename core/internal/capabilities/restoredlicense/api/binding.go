package api

import (
	"regexp"

	restoredlicensev1 "github.com/Onurlulardan/harpy/contracts/restoredlicense/v1"
	"github.com/google/uuid"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Binding struct {
	OperationID            uuid.UUID
	EnvironmentID          uuid.UUID
	RestoreReceiptID       uuid.UUID
	RestoreReceiptSHA256   string
	LicenseBindingID       uuid.UUID
	LicenseID              uuid.UUID
	LicenseeOrganizationID uuid.UUID
	DeliveryKind           string
	Issuer                 string
	PreviousSequence       int64
	PreviousSHA256         string
	NextSequence           int64
	NextSHA256             string
	ManifestSHA256         string
}

func (binding Binding) valid() bool {
	return validV7(binding.OperationID) && binding.EnvironmentID != uuid.Nil &&
		binding.RestoreReceiptID != uuid.Nil && digestPattern.MatchString(binding.RestoreReceiptSHA256) &&
		binding.LicenseBindingID != uuid.Nil && binding.LicenseID != uuid.Nil &&
		binding.LicenseeOrganizationID != uuid.Nil && validDelivery(binding.DeliveryKind) &&
		binding.Issuer != "" && len(binding.Issuer) <= 256 &&
		binding.PreviousSequence > 0 && digestPattern.MatchString(binding.PreviousSHA256) &&
		binding.NextSequence > binding.PreviousSequence && digestPattern.MatchString(binding.NextSHA256) &&
		binding.ManifestSHA256 == restoredlicensev1.ManifestSHA256()
}

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}

func validDelivery(value string) bool {
	return value == "connected" || value == "airgap" || value == "development"
}
