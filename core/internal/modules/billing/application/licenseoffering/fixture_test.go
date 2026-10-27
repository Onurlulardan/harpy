package licenseoffering

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	licensev2 "github.com/Onurlulardan/harpy/contracts/license/v2"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/google/uuid"
)

type offeringFixture struct {
	now                              time.Time
	licenseID, organizationID        uuid.UUID
	environmentID, offeringVersionID uuid.UUID
	document                         []byte
	snapshot                         fixedSnapshot
}

func newOfferingFixture(t *testing.T) offeringFixture {
	t.Helper()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	fixture := offeringFixture{
		now: now, licenseID: uuid.New(), organizationID: uuid.New(),
		environmentID: mustV7(t), offeringVersionID: mustV7(t),
	}
	payload := licensev2.SnapshotPayload{
		Schema: licensev2.SnapshotSchema, Issuer: "harpy",
		Audience: licensev2.SnapshotAudience, Product: licensev2.SnapshotProduct,
		LicenseID:              fixture.licenseID.String(),
		LicenseeOrganizationID: fixture.organizationID.String(),
		EnvironmentID:          fixture.environmentID.String(), DeploymentMode: "onpremise",
		DeliveryKind: licensev2.DeliveryConnected, Sequence: 5,
		Status: licensev2.StatusActive, IssuedAt: instant(now, -2*time.Hour),
		NotBefore: instant(now, -time.Hour), RefreshAfter: instant(now, time.Hour),
		GrantValidUntil: instant(now, 2*time.Hour), GraceUntil: instant(now, 3*time.Hour),
		CommercialExpiresAt: instant(now, 24*time.Hour), ReleaseChannel: "stable",
		Features: licensev2.Features{FormBuilder: true, WorkflowBuilder: true, APIBuilder: true},
		Limits:   licensev2.Limits{MaxUsers: 10, MaxForms: 0, MaxTenants: 1},
	}
	fixture.document = canonicalValue(t, payload)
	digest := sha256.Sum256(fixture.document)
	fixture.snapshot = fixedSnapshot{
		identity: billingapi.LicenseIdentity{
			LicenseID:              fixture.licenseID.String(),
			LicenseeOrganizationID: fixture.organizationID.String(),
			EnvironmentID:          fixture.environmentID.String(), DeliveryKind: licensev2.DeliveryConnected,
		},
		issuer: "harpy", sequence: 5, digest: hex.EncodeToString(digest[:]), channel: "stable",
		projection: billingapi.LicenseEntitlementProjection{
			Boolean: map[string]bool{
				"workflow.builder.enabled": true,
				"api.builder.enabled": true, "analytics.dashboard_builder.enabled": false,
				"workflow.enabled": true, "api.access": true,
			},
			Limits: map[string]billingapi.LicenseLimit{
				"iam.active_users": {Value: 10},
				"workflow.active_definitions": {}, "api.active_endpoints": {},
				"analytics.active_dashboards": {},
				"documents.storage_bytes": {Unlimited: true},
			}, Channel: "stable",
		},
	}
	return fixture
}

func (fixture offeringFixture) checker() billingapi.LicenseSnapshotChecker {
	return fixtureChecker{document: fixture.document, snapshot: fixture.snapshot}
}

func canonicalValue(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		raw, 1<<20,
	)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}

type fixedSnapshot struct {
	identity   billingapi.LicenseIdentity
	issuer     string
	sequence   int64
	digest     string
	channel    string
	projection billingapi.LicenseEntitlementProjection
}

func (snapshot fixedSnapshot) LicenseIdentity() billingapi.LicenseIdentity { return snapshot.identity }
func (snapshot fixedSnapshot) Issuer() string                              { return snapshot.issuer }
func (snapshot fixedSnapshot) Sequence() int64                             { return snapshot.sequence }
func (snapshot fixedSnapshot) Digest() string                              { return snapshot.digest }
func (snapshot fixedSnapshot) Channel() string                             { return snapshot.channel }
func (snapshot fixedSnapshot) OfferingProjection() billingapi.LicenseEntitlementProjection {
	return snapshot.projection
}

type fixtureChecker struct {
	document []byte
	snapshot fixedSnapshot
}

func (checker fixtureChecker) CheckLicenseSnapshot(
	_ context.Context, document []byte, binding billingapi.LicenseVerificationBinding, _ time.Time,
) (billingapi.VerifiedLicenseSnapshot, error) {
	if !bytes.Equal(document, checker.document) || binding != checker.snapshot.identity {
		return nil, errors.New("snapshot rejected")
	}
	return checker.snapshot, nil
}

func instant(base time.Time, offset time.Duration) string {
	return base.Add(offset).Format(time.RFC3339)
}

func mustV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
