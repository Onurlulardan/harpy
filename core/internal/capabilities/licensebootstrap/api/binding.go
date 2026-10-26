package api

import (
	"regexp"
	"time"

	licensebootstrapv1 "github.com/Onurlulardan/harpy/contracts/licensebootstrap/v1"
	"github.com/google/uuid"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Binding struct {
	OperationID            uuid.UUID
	EnvironmentID          uuid.UUID
	LicenseID              uuid.UUID
	LicenseeOrganizationID uuid.UUID
	DeploymentMode         string
	DeliveryKind           string
	Issuer                 string
	Sequence               int64
	SnapshotSHA256         string
	OfferingVersionID      uuid.UUID
	Revision               int64
	OfferingSHA256         string
	EffectiveAt            time.Time
	ManifestSHA256         string
}

func (binding Binding) valid() bool {
	return validV7(binding.OperationID) && validV7(binding.EnvironmentID) &&
		binding.LicenseID != uuid.Nil && binding.LicenseeOrganizationID != uuid.Nil &&
		binding.DeploymentMode == "onpremise" &&
		(binding.DeliveryKind == "connected" || binding.DeliveryKind == "airgap" ||
			binding.DeliveryKind == "development") &&
		binding.Issuer != "" && len(binding.Issuer) <= 256 && binding.Sequence > 0 &&
		digestPattern.MatchString(binding.SnapshotSHA256) &&
		validV7(binding.OfferingVersionID) && binding.Revision > 0 &&
		digestPattern.MatchString(binding.OfferingSHA256) &&
		!binding.EffectiveAt.IsZero() && binding.EffectiveAt.Location() == time.UTC &&
		binding.ManifestSHA256 == licensebootstrapv1.ManifestSHA256()
}

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}
