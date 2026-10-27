package recoveryexecution

import "testing"

func TestUpdateLeaseRequiresPersistedChange(t *testing.T) {
	if err := testService(&executionStoreFixture{changed: true}).UpdateLease(
		t.Context(), testIdentity(),
	); err != nil {
		t.Fatal(err)
	}
}
