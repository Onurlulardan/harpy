package restoreexecution

import "testing"

func TestClaimValidatesExactPersistedEvidence(t *testing.T) {
	claim := validClaimFixture()
	fixture := &storeFixture{claim: claim, found: true}
	result, found, err := New(fixture).Claim(t.Context(), fixtureIdentity, claim.PlanSHA256)
	if err != nil || !found || result.RecoverySetID != claim.RecoverySetID {
		t.Fatalf("Claim() = %#v, %v, %v", result, found, err)
	}
	fixture.claim.ComponentInventory = []byte(`{"database_archive":{}}`)
	if _, _, err = New(fixture).Claim(
		t.Context(), fixtureIdentity, claim.PlanSHA256,
	); err == nil {
		t.Fatal("incomplete component inventory was accepted")
	}
}
