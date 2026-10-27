package maintenance

import "testing"

func TestExecuteProfileBackfillRoutesClosedKind(t *testing.T) {
	executeFixture(t, KindProfileBackfill)
	if t.Failed() {
		t.FailNow()
	}
}
