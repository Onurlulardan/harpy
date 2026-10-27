package recoveryexecution

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestPutComponentsRejectsIncompleteInventory(t *testing.T) {
	fixture := &executionStoreFixture{changed: true}
	err := testService(fixture).PutComponents(
		t.Context(),
		ports.RecoveryComponentBatch{
			RecoveryExecutionIdentity: testIdentity(),
			Components:                []byte(`{}`),
		},
	)
	if err != ErrInvalid || fixture.putCalls != 0 {
		t.Fatalf("error=%v put calls=%d", err, fixture.putCalls)
	}
}
