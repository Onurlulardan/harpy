package recoveryset

import "testing"

func TestRestorePlanAndReceiptRefuseMissingOrAuthorityChangingArtifacts(t *testing.T) {
	request := restoreFixture()
	plan, err := BuildRestorePlan(request, uuidFixture(7), digestFixture("3"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ParseRestorePlan(plan); err != nil {
		t.Fatal(err)
	}
	if _, err = ParseRestorePlan(nil); err == nil {
		t.Fatal("missing restore plan was accepted")
	}
	receipt := RestoreReceiptPayload{Schema: RestoreReceiptSchema,
		OperationID: uuidFixture(6), RecoverySetID: uuidFixture(7),
		PlanSHA256: digestFixture("2"), ManifestSHA256: digestFixture("3"),
		Identity: request.Manifest.Identity, PreviousDatabaseSystemID: uuidFixture(8),
		NewDatabaseSystemID:           uuidFixture(9),
		PreviousDatabaseLocatorSHA256: digestFixture("a"),
		NewDatabaseLocatorSHA256:      digestFixture("b"),
		DrainAckSHA256:                digestFixture("c"), SwitchAckSHA256: digestFixture("d"),
		ObjectMerkleSHA256: digestFixture("4"),
		LocatorRemapSHA256: digestFixture("5"), SafetyEvidenceSHA256: digestFixture("6"),
		SwitchOutcome: "switched_same_environment"}
	encoded, err := BuildRestoreReceipt(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ParseRestoreReceipt(encoded); err != nil {
		t.Fatal(err)
	}
	receipt.OwnerChanged = true
	if _, err = BuildRestoreReceipt(receipt); err == nil {
		t.Fatal("owner-changing restore receipt was accepted")
	}
	receipt.OwnerChanged, receipt.Seeded = false, true
	if _, err = BuildRestoreReceipt(receipt); err == nil {
		t.Fatal("seeded restore receipt was accepted")
	}
}
