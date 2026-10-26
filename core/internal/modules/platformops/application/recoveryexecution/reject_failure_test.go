package recoveryexecution

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestRejectFailureRequiresClosedFailureCode(t *testing.T) {
	fixture := &executionStoreFixture{changed: true}
	failure := ports.RecoveryExecutionFailure{
		RecoveryExecutionIdentity: testIdentity(),
		Code:                      "backup_execution_failed",
	}
	if err := testService(fixture).RejectFailure(t.Context(), failure); err != nil {
		t.Fatal(err)
	}
	failure.Code = "INVALID"
	if err := testService(fixture).RejectFailure(t.Context(), failure); err == nil {
		t.Fatal("invalid failure code was accepted")
	}
}
