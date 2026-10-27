package restoreexecution

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestRejectFailureUsesClosedFailureCode(t *testing.T) {
	record := ports.RestoreFailure{RestoreExecutionIdentity: fixtureIdentity,
		Code: "archive_verification_failed", Fenced: true}
	if err := New(&storeFixture{changed: true}).RejectFailure(t.Context(), record); err != nil {
		t.Fatal(err)
	}
}
