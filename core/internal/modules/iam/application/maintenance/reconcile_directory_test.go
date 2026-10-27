package maintenance

import "testing"

func TestExecuteDirectoryRepairRoutesClosedKind(t *testing.T) {
	executeFixture(t, KindDirectoryRepair)
	if t.Failed() {
		t.FailNow()
	}
}
