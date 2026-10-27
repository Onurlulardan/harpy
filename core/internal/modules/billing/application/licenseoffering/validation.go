package licenseoffering

import (
	"regexp"
	"time"

	licensebootstrap "github.com/Onurlulardan/harpy/core/internal/capabilities/licensebootstrap/api"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/google/uuid"
)

const timeFormat = time.RFC3339Nano

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validPlan(plan Plan) bool {
	return plan.OfferingVersionID != uuid.Nil &&
		plan.OfferingVersionID.Version() == uuid.Version(7) && plan.Revision > 0 &&
		!plan.EffectiveAt.IsZero() && plan.EffectiveAt.Location() == time.UTC
}

func snapshotMatches(
	binding licensebootstrap.Binding,
	snapshot billingapi.VerifiedLicenseSnapshot,
) bool {
	identity := snapshot.LicenseIdentity()
	return identity.LicenseID == binding.LicenseID.String() &&
		identity.LicenseeOrganizationID == binding.LicenseeOrganizationID.String() &&
		identity.EnvironmentID == binding.EnvironmentID.String() &&
		string(identity.DeliveryKind) == binding.DeliveryKind &&
		snapshot.Issuer() == binding.Issuer && snapshot.Sequence() == binding.Sequence &&
		snapshot.Digest() == binding.SnapshotSHA256
}

func validReceipt(binding licensebootstrap.Binding, receipt Receipt) bool {
	return receipt.OperationID == binding.OperationID &&
		receipt.OfferingVersionID == binding.OfferingVersionID &&
		receipt.Revision == binding.Revision &&
		receipt.OfferingSHA256 == binding.OfferingSHA256 &&
		digestPattern.MatchString(receipt.OfferingSHA256)
}
