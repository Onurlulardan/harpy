package restoreexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestImportReceiptRequiresExactComponentInventory(t *testing.T) {
	document := []byte("signed receipt")
	digest := sha256.Sum256(document)
	record := ports.RestoreReceiptImport{
		OperationID:   fixtureIdentity.OperationID,
		RecoverySetID: "44444444-4444-4444-8444-444444444444",
		EnvironmentID: "55555555-5555-4555-8555-555555555555",
		PlanSHA256:    digestFixture(), ManifestSHA256: digestFixture(),
		ReceiptSHA256: hex.EncodeToString(digest[:]), ReceiptBytes: document,
		PreviousDatabaseSystemID: "77777777-7777-4777-8777-777777777777",
		NewDatabaseSystemID:      "88888888-8888-4888-8888-888888888888",
		ObjectMerkleSHA256:       digestFixture(), LocatorRemapSHA256: digestFixture(),
		SafetyEvidenceSHA256: digestFixture(), DrainAckSHA256: digestFixture(),
		SwitchAckSHA256:              digestFixture(),
		ControlDatabaseLocatorSHA256: digestFixture(),
		ScratchDatabaseLocatorSHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ComponentInventory:           componentFixture(),
	}
	if err := New(&storeFixture{changed: true}).ImportReceipt(
		t.Context(), record,
	); err != nil {
		t.Fatal(err)
	}
}
