package restoreexecution

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestValidateRestoreRequiresArchiveAndDatabaseEvidence(t *testing.T) {
	record := ports.RestoreVerification{RestoreExecutionIdentity: fixtureIdentity,
		ArchiveReadbackSHA256: digestFixture(),
		NewDatabaseSystemID:   "88888888-8888-4888-8888-888888888888"}
	if err := New(&storeFixture{changed: true}).ValidateRestore(
		t.Context(), record,
	); err != nil {
		t.Fatal(err)
	}
}
