package restoreexecution

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestExecuteApplyRequiresDrainAckDigest(t *testing.T) {
	service := New(&storeFixture{changed: true})
	record := ports.RestoreApplyEvidence{
		RestoreExecutionIdentity:     fixtureIdentity,
		DrainAckSHA256:               digestFixture(),
		ControlDatabaseLocatorSHA256: digestFixture(),
		ScratchDatabaseLocatorSHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
	if err := service.ExecuteApply(t.Context(), record); err != nil {
		t.Fatal(err)
	}
	record.DrainAckSHA256 = "invalid"
	if err := service.ExecuteApply(t.Context(), record); err == nil {
		t.Fatal("invalid drain acknowledgement was accepted")
	}
}
