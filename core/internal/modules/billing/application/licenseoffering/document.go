package licenseoffering

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	licensev2 "github.com/Onurlulardan/harpy/contracts/license/v2"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
)

func BuildDocument(
	plan Plan,
	snapshot billingapi.VerifiedLicenseSnapshot,
) (Document, error) {
	identity := snapshot.LicenseIdentity()
	if !validPlan(plan) || identity.LicenseID == "" || snapshot.Digest() == "" {
		return Document{}, ErrProjection
	}
	projection := snapshot.OfferingProjection()
	booleans := make([]licensev2.BooleanGrant, 0, len(projection.Boolean))
	for key, value := range projection.Boolean {
		booleans = append(booleans, licensev2.BooleanGrant{Key: key, Value: value})
	}
	sort.Slice(booleans, func(left, right int) bool {
		return booleans[left].Key < booleans[right].Key
	})
	limits := make([]licensev2.LimitGrant, 0, len(projection.Limits))
	for key, value := range projection.Limits {
		limits = append(limits, licensev2.LimitGrant{
			Key: key, Value: value.Value, Unlimited: value.Unlimited,
		})
	}
	sort.Slice(limits, func(left, right int) bool { return limits[left].Key < limits[right].Key })
	wire := licensev2.DerivedOffering{
		Schema:            licensev2.DerivedOfferingSchema,
		OfferingVersionID: plan.OfferingVersionID.String(), Revision: plan.Revision,
		EffectiveAt:            plan.EffectiveAt.Format(timeFormat),
		LicenseID:              identity.LicenseID,
		LicenseeOrganizationID: identity.LicenseeOrganizationID,
		EnvironmentID:          identity.EnvironmentID, Issuer: snapshot.Issuer(),
		SnapshotSequence: snapshot.Sequence(), SnapshotSHA256: snapshot.Digest(),
		ReleaseChannel: snapshot.Channel(), Booleans: booleans, Limits: limits,
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		return Document{}, ErrProjection
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		raw, 1<<20,
	)
	if err != nil {
		return Document{}, ErrProjection
	}
	digest := sha256.Sum256(canonical)
	return Document{
		Bytes: canonical, SHA256: hex.EncodeToString(digest[:]), Projection: projection,
	}, nil
}
