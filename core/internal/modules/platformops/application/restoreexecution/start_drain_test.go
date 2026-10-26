package restoreexecution

import "testing"

func TestStartDrainRequiresClaimIdentity(t *testing.T) {
	if err := New(&storeFixture{changed: true}).StartDrain(
		t.Context(), fixtureIdentity,
	); err != nil {
		t.Fatal(err)
	}
}
