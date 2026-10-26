package licenseauthority

import (
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/google/uuid"
)

type VerifiedSnapshot struct {
	licenseID, organizationID, environmentID uuid.UUID
	issuer, keyID, digest, releaseChannel    string
	delivery                                 DeliveryKind
	status                                   Status
	sequence                                 int64
	issuedAt, notBefore, refreshAfter        time.Time
	grantValidUntil, graceUntil              time.Time
	commercialExpiresAt                      time.Time
	features                                 FeatureClaims
	limits                                   LimitClaims
	canonicalPayload, signedDocument         []byte
}

func (snapshot VerifiedSnapshot) LicenseID() uuid.UUID       { return snapshot.licenseID }
func (snapshot VerifiedSnapshot) OrganizationID() uuid.UUID  { return snapshot.organizationID }
func (snapshot VerifiedSnapshot) EnvironmentID() uuid.UUID   { return snapshot.environmentID }
func (snapshot VerifiedSnapshot) Issuer() string             { return snapshot.issuer }
func (snapshot VerifiedSnapshot) KeyID() string              { return snapshot.keyID }
func (snapshot VerifiedSnapshot) Digest() string             { return snapshot.digest }
func (snapshot VerifiedSnapshot) DeliveryKind() DeliveryKind { return snapshot.delivery }
func (snapshot VerifiedSnapshot) Status() Status             { return snapshot.status }
func (snapshot VerifiedSnapshot) Sequence() int64            { return snapshot.sequence }
func (snapshot VerifiedSnapshot) IssuedAt() time.Time        { return snapshot.issuedAt }
func (snapshot VerifiedSnapshot) NotBefore() time.Time       { return snapshot.notBefore }
func (snapshot VerifiedSnapshot) RefreshAfter() time.Time    { return snapshot.refreshAfter }
func (snapshot VerifiedSnapshot) LicenseGrantDeadline() time.Time {
	return snapshot.grantValidUntil
}
func (snapshot VerifiedSnapshot) GraceUntil() time.Time { return snapshot.graceUntil }
func (snapshot VerifiedSnapshot) CommercialExpiresAt() time.Time {
	return snapshot.commercialExpiresAt
}
func (snapshot VerifiedSnapshot) Channel() string         { return snapshot.releaseChannel }
func (snapshot VerifiedSnapshot) Features() FeatureClaims { return snapshot.features }
func (snapshot VerifiedSnapshot) Limits() LimitClaims     { return snapshot.limits }

func (snapshot VerifiedSnapshot) LicenseIdentity() billingapi.LicenseIdentity {
	return billingapi.LicenseIdentity{
		LicenseID:              snapshot.licenseID.String(),
		LicenseeOrganizationID: snapshot.organizationID.String(),
		EnvironmentID:          snapshot.environmentID.String(), DeliveryKind: snapshot.delivery,
	}
}

func (snapshot VerifiedSnapshot) OfferingProjection() billingapi.LicenseEntitlementProjection {
	return ProjectEntitlements(snapshot)
}

func (snapshot VerifiedSnapshot) CanonicalPayload() []byte {
	return append([]byte(nil), snapshot.canonicalPayload...)
}

func (snapshot VerifiedSnapshot) SignedDocument() []byte {
	return append([]byte(nil), snapshot.signedDocument...)
}

func (snapshot VerifiedSnapshot) RuntimeState(now time.Time) RuntimeState {
	if snapshot.status != StatusActive || now.After(snapshot.graceUntil) {
		return RuntimeBuildOff
	}
	if now.After(snapshot.grantValidUntil) {
		return RuntimeGrace
	}
	return RuntimeValid
}
