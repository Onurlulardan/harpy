package v1adoption

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestEvaluateRejectsEvidenceOutsideExactTenantInventory(t *testing.T) {
	t.Parallel()
	tenantID := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	unknownID := uuid.MustParse("99999999-9999-7999-8999-999999999999")
	inventory := validInventory(
		ModeOnPremise,
		tenantID,
		uuid.MustParse("22222222-2222-7222-8222-222222222222"),
	)
	inventory.Protected = append(inventory.Protected, ProtectedAccountEvidence{
		TenantID: unknownID,
		UserID:   uuid.MustParse("88888888-8888-7888-8888-888888888888"),
	})
	inventory.Billing = []BillingEvidence{
		validLicenseEvidence(tenantID),
		validLicenseEvidence(unknownID),
	}

	_, err := Evaluate(inventory)
	assertIssueEvidence(t, err, IssueInventoryInvalid, unknownID, "protected_account")
	assertIssueEvidence(t, err, IssueInventoryInvalid, unknownID, "billing")
}

func TestEvaluateFencesMissingOwnerMappingAndCredential(t *testing.T) {
	t.Parallel()
	tenantID := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	inventory := validInventory(
		ModeOnPremise,
		tenantID,
		uuid.MustParse("22222222-2222-7222-8222-222222222222"),
	)
	inventory.Protected[0].CanonicalIAMMapped = false
	inventory.Protected[0].UsableCredential = false
	inventory.Billing = []BillingEvidence{validLicenseEvidence(tenantID)}

	_, err := Evaluate(inventory)
	assertIssue(t, err, IssueOwnerMappingMissing)
	assertIssue(t, err, IssueOwnerCredentialMissing)
}

func assertIssueEvidence(
	t *testing.T,
	err error,
	code string,
	tenantID uuid.UUID,
	evidence string,
) {
	t.Helper()
	var fenced *FenceError
	if !errors.As(err, &fenced) {
		t.Fatalf("error = %v", err)
	}
	for _, issue := range fenced.Issues {
		if issue.Code == code && issue.Tenant == tenantID.String() &&
			issue.Evidence == evidence {
			return
		}
	}
	t.Fatalf("missing %s/%s/%s in %#v", code, tenantID, evidence, fenced.Issues)
}

func validLicenseEvidence(tenantID uuid.UUID) BillingEvidence {
	return BillingEvidence{
		TenantID: tenantID, Kind: BillingLicenseDerived,
		VerifiedV2Publication: true,
		Offering:              validOffering("license_enterprise_v1", 7),
	}
}
