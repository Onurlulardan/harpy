package v1adoption

import (
	"testing"

	"github.com/google/uuid"
)

func TestEvaluateFencesRawOnPremiseLicenseEvidence(t *testing.T) {
	t.Parallel()
	tenantID := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	inventory := validInventory(
		ModeOnPremise,
		tenantID,
		uuid.MustParse("22222222-2222-7222-8222-222222222222"),
	)
	inventory.Billing = []BillingEvidence{{
		TenantID: tenantID, Kind: BillingLicenseDerived,
		Offering: validOffering("legacy_raw_license", 0),
	}}

	_, err := Evaluate(inventory)
	assertIssue(t, err, IssueLicenseExchangeRequired)
}

func TestEvaluateFencesInexactSaaSBaseEvidence(t *testing.T) {
	t.Parallel()
	control := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	customer := uuid.MustParse("33333333-3333-7333-8333-333333333333")
	cases := []struct {
		name     string
		evidence BillingEvidence
	}{
		{
			name: "control offering",
			evidence: BillingEvidence{
				TenantID: control, Kind: BillingSaaSControl,
				VerifiedV2Publication: true,
				Offering:              validOffering("wrong-control", 0),
			},
		},
		{
			name: "free subscription evidence",
			evidence: BillingEvidence{
				TenantID: customer, Kind: BillingSaaSFree,
				VerifiedV2Publication:      true,
				SubscriptionEvidenceSHA256: hash64("e"),
				Offering:                   validOffering(SaaSFreeOfferingKey, 0),
			},
		},
		{
			name: "paid without subscription evidence",
			evidence: BillingEvidence{
				TenantID: customer, Kind: BillingSaaSPaid,
				VerifiedV2Publication: true,
				Offering:              validOffering("enterprise_v4", 0),
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			inventory := validSaaSInventory(control, customer)
			for index := range inventory.Billing {
				if inventory.Billing[index].TenantID == test.evidence.TenantID {
					inventory.Billing[index] = test.evidence
				}
			}
			_, err := Evaluate(inventory)
			assertIssue(t, err, IssueBaseEvidenceAmbiguous)
		})
	}
}

func validSaaSInventory(control, customer uuid.UUID) Inventory {
	inventory := validInventory(
		ModeSaaS,
		control,
		uuid.MustParse("22222222-2222-7222-8222-222222222222"),
	)
	inventory.Tenants[0].ControlCandidate = true
	inventory.Tenants = append(inventory.Tenants, TenantEvidence{
		TenantID: customer, Status: "ACTIVE",
	})
	inventory.Billing = []BillingEvidence{
		{
			TenantID: control, Kind: BillingSaaSControl,
			VerifiedV2Publication: true,
			Offering:              validOffering(SaaSControlOfferingKey, 0),
		},
		{
			TenantID: customer, Kind: BillingSaaSFree,
			VerifiedV2Publication: true,
			Offering:              validOffering(SaaSFreeOfferingKey, 0),
		},
	}
	return inventory
}
