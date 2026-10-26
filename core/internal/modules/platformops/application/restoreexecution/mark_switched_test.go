package restoreexecution

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestMarkSwitchedRequiresAllVerificationDigests(t *testing.T) {
	record := ports.RestoreSwitchEvidence{RestoreExecutionIdentity: fixtureIdentity,
		ObjectMerkleSHA256: digestFixture(), LocatorRemapSHA256: digestFixture(),
		SafetyEvidenceSHA256: digestFixture(), SwitchAckSHA256: digestFixture(),
		PreviousDatabaseSystemID:     "77777777-7777-4777-8777-777777777777",
		NewDatabaseSystemID:          "88888888-8888-4888-8888-888888888888",
		ControlDatabaseLocatorSHA256: digestFixture(),
		ScratchDatabaseLocatorSHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
	if err := New(&storeFixture{changed: true}).MarkSwitched(
		t.Context(), record,
	); err != nil {
		t.Fatal(err)
	}
}
