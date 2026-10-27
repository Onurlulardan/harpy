package v1adoption

import (
	"testing"

	"github.com/google/uuid"
)

func TestEvaluateFencesOnPremiseControlTenantEvidence(t *testing.T) {
	t.Parallel()
	tenantID := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	inventory := validInventory(
		ModeOnPremise,
		tenantID,
		uuid.MustParse("22222222-2222-7222-8222-222222222222"),
	)
	inventory.Tenants[0].ControlCandidate = true
	inventory.Billing = []BillingEvidence{{
		TenantID:              tenantID,
		Kind:                  BillingLicenseDerived,
		VerifiedV2Publication: true,
		Offering:              validOffering("license_enterprise_v1", 7),
	}}

	_, err := Evaluate(inventory)
	assertIssue(t, err, IssueControlModeMismatch)
}
