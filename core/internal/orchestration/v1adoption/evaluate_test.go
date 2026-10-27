package v1adoption

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEvaluateOnPremiseMapsSeparateOwnerAndProtectedEvidence(t *testing.T) {
	t.Parallel()
	tenantID := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-7222-8222-222222222222")
	inventory := validInventory(ModeOnPremise, tenantID, userID)
	inventory.Billing = []BillingEvidence{{
		TenantID: tenantID, Kind: BillingLicenseDerived, VerifiedV2Publication: true,
		Offering: validOffering("license_enterprise_v1", 7),
	}}
	plan, err := Evaluate(inventory)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PrimaryTenantID != tenantID || plan.OwnerUserID != userID ||
		plan.OwnerAuthorizationVersion != 7 ||
		plan.TenantBindings[0].Purpose != "onpremise_primary" ||
		plan.BaseAssignments[0].Kind != BillingLicenseDerived ||
		!validSHA(plan.InventorySHA256) || !validSHA(plan.PlanSHA256) {
		t.Fatalf("plan = %#v", plan)
	}
}

func TestEvaluateFencesMultiTenantOnPremiseAndRawLicense(t *testing.T) {
	t.Parallel()
	tenantID := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	inventory := validInventory(ModeOnPremise, tenantID,
		uuid.MustParse("22222222-2222-7222-8222-222222222222"))
	inventory.Tenants = append(inventory.Tenants, TenantEvidence{
		TenantID: uuid.MustParse("33333333-3333-7333-8333-333333333333"), Status: "ACTIVE",
	})
	inventory.Billing = []BillingEvidence{{
		TenantID: tenantID, Kind: BillingLicenseDerived,
		Offering: validOffering("legacy_raw_license", 0),
	}}
	_, err := Evaluate(inventory)
	assertIssue(t, err, IssueOnPremTenantAmbiguous)
}

func TestEvaluateSaaSMapsControlFreeAndPaidExactly(t *testing.T) {
	t.Parallel()
	control := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	free := uuid.MustParse("33333333-3333-7333-8333-333333333333")
	paid := uuid.MustParse("44444444-4444-7444-8444-444444444444")
	inventory := validInventory(ModeSaaS, control,
		uuid.MustParse("22222222-2222-7222-8222-222222222222"))
	inventory.Tenants[0].ControlCandidate = true
	inventory.Tenants = append(inventory.Tenants,
		TenantEvidence{TenantID: free, Status: "ACTIVE"},
		TenantEvidence{TenantID: paid, Status: "ACTIVE"})
	inventory.Billing = []BillingEvidence{
		{TenantID: control, Kind: BillingSaaSControl, VerifiedV2Publication: true,
			Offering: validOffering(SaaSControlOfferingKey, 0)},
		{TenantID: free, Kind: BillingSaaSFree, VerifiedV2Publication: true,
			Offering: validOffering(SaaSFreeOfferingKey, 0)},
		{TenantID: paid, Kind: BillingSaaSPaid, VerifiedV2Publication: true,
			SubscriptionEvidenceSHA256: hash64("b"), Offering: validOffering("enterprise_v4", 0)},
	}
	plan, err := Evaluate(inventory)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.TenantBindings) != 3 || len(plan.BaseAssignments) != 3 ||
		plan.TenantBindings[0].Purpose != "saas_control" ||
		plan.BaseAssignments[2].SubscriptionEvidenceSHA256 != hash64("b") {
		t.Fatalf("plan = %#v", plan)
	}
}

func TestEvaluateFencesAmbiguousSaaSControlAndOwner(t *testing.T) {
	t.Parallel()
	control := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	inventory := validInventory(ModeSaaS, control,
		uuid.MustParse("22222222-2222-7222-8222-222222222222"))
	inventory.Protected = append(inventory.Protected, inventory.Protected[0])
	_, err := Evaluate(inventory)
	assertIssue(t, err, IssueControlTenantAmbiguous)
	assertIssue(t, err, IssueProtectedOwnerAmbiguous)
}

func assertIssue(t *testing.T, err error, code string) {
	t.Helper()
	if !errors.Is(err, ErrFenced) {
		t.Fatalf("error = %v", err)
	}
	var fenced *FenceError
	if !errors.As(err, &fenced) {
		t.Fatalf("error type = %T", err)
	}
	for _, issue := range fenced.Issues {
		if issue.Code == code {
			return
		}
	}
	t.Fatalf("missing issue %s in %#v", code, fenced.Issues)
}

func validInventory(mode DeploymentMode, tenantID, userID uuid.UUID) Inventory {
	return Inventory{
		EnvironmentID: uuid.MustParse("aaaaaaaa-aaaa-7aaa-8aaa-aaaaaaaaaaaa"), Mode: mode,
		Tenants: []TenantEvidence{{TenantID: tenantID, Status: "ACTIVE"}},
		Protected: []ProtectedAccountEvidence{{
			TenantID: tenantID, UserID: userID, Active: true,
			CanonicalIAMMapped: true, ProtectedAdminBound: true,
			UsableCredential: true, AuthorizationVersion: 7,
			MappingSHA256: hash64("a"),
		}},
	}
}

func validOffering(key string, sourceSequence int64) OfferingPin {
	pin := OfferingPin{
		Key: key, VersionID: uuid.MustParse("55555555-5555-7555-8555-555555555555"),
		Revision: 1, SHA256: hash64("c"), EffectiveAt: time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC),
		SourceSequence: sourceSequence,
	}
	if sourceSequence > 0 {
		pin.SourceSHA256 = hash64("d")
	}
	return pin
}

func hash64(value string) string {
	result := ""
	for len(result) < 64 {
		result += value
	}
	return result[:64]
}
