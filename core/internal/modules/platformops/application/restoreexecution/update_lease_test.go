package restoreexecution

import "testing"

func TestUpdateLeaseRequiresPersistedChange(t *testing.T) {
	if err := New(&storeFixture{changed: true}).UpdateLease(
		t.Context(), fixtureIdentity,
	); err != nil {
		t.Fatal(err)
	}
}
