package maintenance

import "testing"

func TestClaimRejectsReferenceDigestOrTenantSubstitution(t *testing.T) {
	claim := fixtureClaim(t, KindDirectoryRepair)
	claim.PayloadSHA256 = digestReference(claim.PayloadReference + ":tampered")
	if claim.Validate() == nil {
		t.Fatal("tampered payload digest was accepted")
	}
	claim = fixtureClaim(t, KindProfileBackfill)
	claim.TenantID = fixtureJobID
	if claim.Validate() == nil {
		t.Fatal("cross-tenant profile reference was accepted")
	}
}
