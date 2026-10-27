package maintenance

import "testing"

func TestExecuteIdentifierRetireRoutesClosedKind(t *testing.T) {
	executeFixture(t, KindIdentifierRetire)
	if t.Failed() {
		t.FailNow()
	}
}
