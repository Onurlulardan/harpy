package maintenance

import "testing"

func TestExecuteDirectoryTombstonePurgeRoutesClosedKind(t *testing.T) {
	executeFixture(t, KindDirectoryTombstonePurge)
	if t.Failed() {
		t.FailNow()
	}
}
